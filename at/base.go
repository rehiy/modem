package at

import (
	"bufio"
	"context"
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
	PortName        string           // 串口名称，如 '/dev/ttyUSB0' 或 'COM3'
	BaudRate        int              // 波特率，如 115200
	ReadTimeout     time.Duration    // 读取超时时间
	Parity          byte             // 校验位，如 'N', 'E', 'O'
	StopBits        byte             // 停止位，如 1, 2
	CommandSet      *CommandSet      // 自定义 AT 命令集，如果为 nil 则使用默认命令集
	NotificationSet *NotificationSet // 自定义通知类型集，如果为 nil 则使用默认通知集
	ResponseSet     *ResponseSet     // 自定义响应类型集，如果为 nil 则使用默认响应集
}

// Modem 连接
type Device struct {
	port          *serial.Port
	config        Config
	commands      CommandSet      // 使用的 AT 命令集
	notifications NotificationSet // 使用的通知类型集
	responses     ResponseSet     // 使用的响应类型集
	isClosed      atomic.Bool     // 连接是否已关闭（原子操作保证并发安全）

	// 统一读取相关字段
	reader       *bufio.Reader       // 统一的读取器
	responseChan chan string         // 命令响应通道
	urcHandler   NotificationHandler // 通知处理函数
	urcMu        sync.RWMutex        // 保护通知函数的读写锁
	mu           sync.Mutex          // 保护命令发送的互斥锁
}

type NotificationHandler func(notification string)

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
		config:        config,
		commands:      commands,
		notifications: notifications,
		responses:     responses,
		reader:        bufio.NewReader(port),
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
		return nil, ErrDeviceClosed
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

// ListenNotifications 注册 modem 通知处理器
func (m *Device) ListenNotifications(handler NotificationHandler) (error, context.CancelFunc) {
	if m.isClosed.Load() {
		return ErrDeviceClosed, nil
	}

	// 注册通知处理器
	m.urcMu.Lock()
	m.urcHandler = handler
	m.urcMu.Unlock()

	// 监听上下文取消，取消时移除处理器
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		m.urcMu.Lock()
		m.urcHandler = nil
		m.urcMu.Unlock()
	}()

	return nil, cancel
}

// ===== 原生读写 =====

// readLoop 从串口读取数据并分发
func (m *Device) readLoop() {
	for !m.isClosed.Load() {
		line, err := m.reader.ReadString('\n')
		if err != nil {
			// 连接已关闭，退出循环
			if m.isClosed.Load() {
				return
			}
			// EOF 表示连接已断开，应该退出循环
			if err == io.EOF {
				_ = m.Close()
				return
			}
			// 其他错误，继续监听
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue // 忽略空行
		}

		// 分发数据
		if m.notifications.IsNotification(line) {
			m.urcMu.RLock()
			handler := m.urcHandler
			m.urcMu.RUnlock()
			if handler != nil {
				go handler(line)
			}
		} else {
			select {
			case m.responseChan <- line:
			default:
				// 通道满了，丢弃数据（避免阻塞）
				log.Printf("discarding data: %s", line)
			}
		}
	}
}

// writeString 写入数据到串口
func (m *Device) writeString(data string) error {
	if m.isClosed.Load() {
		return ErrDeviceClosed
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
