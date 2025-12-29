package at

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tarm/serial"
)

// Modem 配置
type Config struct {
	// 串口参数
	PortName    string        // 串口名称，如 '/dev/ttyUSB0' 或 'COM3'
	BaudRate    int           // 波特率，如 115200
	ReadTimeout time.Duration // 超时时间
	Parity      byte          // 校验位，如 'N', 'E', 'O'
	StopBits    byte          // 停止位，如 1, 2
	// 设备参数
	CommandSet      *CommandSet      // 自定义 AT 命令集，如果为 nil 则使用默认命令集
	ResponseSet     *ResponseSet     // 自定义响应类型集，如果为 nil 则使用默认响应集
	NotificationSet *NotificationSet // 自定义通知类型集，如果为 nil 则使用默认通知集
	Notification    func(s string)   // 通知处理函数
}

// Modem 连接
type Device struct {
	port    *serial.Port  // 串口连接
	timeout time.Duration // 超时时间

	commands      CommandSet      // 使用的 AT 命令集
	notifications NotificationSet // 使用的通知类型集
	responses     ResponseSet     // 使用的响应类型集

	responseChan chan string    // 命令响应通道
	urcHandler   func(s string) // 通知处理函数
	isClosed     atomic.Bool    // 连接是否已关闭（原子操作保证并发安全）
	mu           sync.Mutex     // 保护命令发送的互斥锁
}

// New 创建一个新的 AT 连接
func New(config Config) (*Device, error) {
	port, err := serial.OpenPort(&serial.Config{
		Name:        config.PortName,
		Baud:        config.BaudRate,
		ReadTimeout: config.ReadTimeout,
		Parity:      serial.Parity(config.Parity),
		StopBits:    serial.StopBits(config.StopBits),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	// timeout 需要大于 100ms
	timeout := config.ReadTimeout + 100*time.Millisecond

	// 如果没有指定命令集，使用默认命令集
	commands := DefaultCommandSet()
	if config.CommandSet != nil {
		commands = *config.CommandSet
	}

	// 如果没有指定通知集，使用默认通知集
	notifications := DefaultNotificationSet()
	if config.NotificationSet != nil {
		notifications = *config.NotificationSet
	}

	// 如果没有指定响应集，使用默认响应集
	responses := DefaultResponseSet()
	if config.ResponseSet != nil {
		responses = *config.ResponseSet
	}

	dev := &Device{
		port:          port,
		timeout:       timeout,
		commands:      commands,
		notifications: notifications,
		responses:     responses,
		responseChan:  make(chan string, 100), // 带缓冲的通道
	}

	// 启动统一读取循环
	go dev.readLoop()

	return dev, nil
}

// IsLive 检查连接状态
func (m *Device) IsLive() bool {
	return !m.isClosed.Load()
}

// Close 关闭连接
func (m *Device) Close() error {
	if m.isClosed.Swap(true) {
		return nil // 已经关闭过了
	}

	// 关闭响应通道
	close(m.responseChan)

	return m.port.Close()
}

// SendCommand 发送 AT 命令并等待响应
func (m *Device) SendCommand(command string) ([]string, error) {
	if m.isClosed.Load() {
		return nil, fmt.Errorf("device is closed")
	}

	// 添加回车换行符并写入命令
	if err := m.writeString(command + "\r\n"); err != nil {
		return nil, err
	}

	return m.readResponse()
}

// SendCommandExpect 发送 AT 命令并期望特定响应
func (m *Device) SendCommandExpect(command string, expected string) error {
	responses, err := m.SendCommand(command)
	if err != nil {
		return err
	}

	// 检查是否包含期望的响应
	for _, response := range responses {
		if strings.Contains(response, expected) {
			return nil
		}
	}

	return fmt.Errorf("expected response %q not found in %v", expected, responses)
}

// readResponse 从响应通道读取响应
func (m *Device) readResponse() ([]string, error) {
	var responses []string
	timeout := time.After(m.timeout)

	for {
		select {
		case line, ok := <-m.responseChan:
			if !ok {
				return responses, fmt.Errorf("device is closed")
			}

			responses = append(responses, line)
			if m.responses.IsFinalResponse(line) {
				return responses, nil
			}

		case <-timeout:
			return responses, fmt.Errorf("command timeout")
		}
	}
}

// ===== 原生读写 =====

// readLoop 从串口读取数据并分发
func (m *Device) readLoop() {
	reader := bufio.NewReader(m.port)
	for {
		if m.isClosed.Load() {
			return
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			// EOF 表示连接已断开，应该退出循环
			if err == io.EOF {
				_ = m.Close()
				return
			}
			// 其他错误，继续监听
			time.Sleep(time.Second)
			continue
		}

		// 去除空白字符
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理通知消息
		if m.notifications.IsNotification(line) {
			if m.urcHandler != nil {
				go m.urcHandler(line)
			}
			continue
		}

		// 将数据写入响应通道
		select {
		case m.responseChan <- line:
		default:
			// 通道满了，丢弃数据（避免阻塞）
			log.Printf("discarding data: %s", line)
		}
	}
}

// writeString 写入数据到串口
func (m *Device) writeString(data string) error {
	if m.isClosed.Load() {
		return fmt.Errorf("device is closed")
	}

	// 防止并发写
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空响应通道中的旧数据
	for len(m.responseChan) > 0 {
		<-m.responseChan
	}

	n, err := m.port.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("failed to write to port: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("incomplete write: wrote %d of %d bytes", n, len(data))
	}

	return nil
}
