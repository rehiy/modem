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
)

// 端口接口
type Port interface {
	Read(buf []byte) (int, error)   // 读取数据
	Write(data []byte) (int, error) // 写入数据
	Flush() error                   // 刷新缓冲区，暂未使用
	Close() error                   // 关闭连接
}

// 配置参数
type Config struct {
	Timeout         time.Duration        // 超时时间
	CommandSet      *CommandSet          // 自定义 AT 命令集，如果为 nil 则使用默认命令集
	ResponseSet     *ResponseSet         // 自定义响应类型集，如果为 nil 则使用默认响应集
	NotificationSet *NotificationSet     // 自定义通知类型集，如果为 nil 则使用默认通知集
	Printf          func(string, ...any) // 日志输出函数，如果为 nil 则使用 log.Printf
}

// 设备连接
type Device struct {
	port          Port                 // 串口连接
	timeout       time.Duration        // 超时时间
	commands      CommandSet           // 使用的 AT 命令集
	responses     ResponseSet          // 使用的响应类型集
	responseChan  chan string          // 命令响应通道
	notifications NotificationSet      // 使用的通知类型集
	urcHandler    UrcHandler           // 通知处理函数
	printf        func(string, ...any) // 日志输出函数
	closed        atomic.Bool          // 连接是否已关闭（原子操作保证并发安全）
	cmd           atomic.Value         // 当前正在执行的命令
	mu            sync.Mutex           // 保护命令发送的互斥锁
}

// 通知处理函数
type UrcHandler func(string, map[int]string)

// New 创建一个新的设备连接实例
func New(port Port, handler UrcHandler, config *Config) *Device {
	if config == nil {
		config = &Config{}
	}
	if config.Timeout == 0 {
		config.Timeout = time.Second
	}
	if config.CommandSet == nil {
		config.CommandSet = DefaultCommandSet()
	}
	if config.ResponseSet == nil {
		config.ResponseSet = DefaultResponseSet()
	}
	if config.NotificationSet == nil {
		config.NotificationSet = DefaultNotificationSet()
	}
	if config.Printf == nil {
		config.Printf = log.Printf
	}

	dev := &Device{
		port:          port,
		timeout:       config.Timeout,
		commands:      *config.CommandSet,
		responses:     *config.ResponseSet,
		responseChan:  make(chan string, 100),
		notifications: *config.NotificationSet,
		urcHandler:    handler,
		printf:        config.Printf,
	}

	// 开始读取循环
	go dev.readAndDispatch()

	return dev
}

// IsOpen 链接状态
func (m *Device) IsOpen() bool {
	return !m.closed.Load()
}

// Close 关闭连接
func (m *Device) Close() error {
	m.printf("closing device")
	if m.closed.Swap(true) {
		return nil // 已经关闭过了
	}

	close(m.responseChan)
	return m.port.Close()
}

// SendCommand 发送命令并等待响应
func (m *Device) SendCommand(cmd string) ([]string, error) {
	if m.closed.Load() {
		return nil, fmt.Errorf("device closed")
	}

	// 加锁保护
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空响应通道，避免收到残留响应
	for len(m.responseChan) > 0 {
		<-m.responseChan
	}

	// 检查命令是否已包含结束符，避免重复添加
	if !hasTerminator(cmd) {
		cmd = cmd + "\r\n"
	}

	// 记录正在执行的命令
	m.cmd.Store(cmd)
	defer m.cmd.Store("")

	// 向串口写入命令
	if err := m.writeString(cmd); err != nil {
		return nil, err
	}

	return m.readResponse()
}

// SendCommandExpect 发送命令并期望特定响应
func (m *Device) SendCommandExpect(cmd string, expected string) error {
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return err
	}

	// 检查是否包含期望的响应
	for _, response := range responses {
		if strings.Contains(response, expected) {
			return nil
		}
	}
	return fmt.Errorf("%q not found in %v", expected, responses)
}

// SimpleQuery 通用简单信息查询函数
func (m *Device) SimpleQuery(cmd string) (string, error) {
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return "", err
	}

	// 查找不以AT开头的响应
	for _, line := range responses {
		if !strings.HasPrefix(line, "AT") {
			return line, nil
		}
	}
	return "", fmt.Errorf("no info found for %s", cmd)
}

// readResponse 从响应通道读取响应
func (m *Device) readResponse() ([]string, error) {
	var responses []string
	timeout := time.After(m.timeout)

	for {
		select {
		case line, ok := <-m.responseChan:
			if !ok {
				return responses, fmt.Errorf("device closed")
			}
			// 遇到终止响应，返回积累的行
			responses = append(responses, line)
			if m.responses.IsFinal(line) {
				return responses, nil
			}

		case <-timeout:
			return responses, fmt.Errorf("command timeout")
		}
	}
}

// ===== 原生读写 =====

// readAndDispatch 从串口读取数据并分发
func (m *Device) readAndDispatch() {
	reader := bufio.NewReader(m.port)
	for {
		if m.closed.Load() {
			return
		}

		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				m.printf("read error: %v", err)
			}
			time.Sleep(m.timeout / 2)
			continue
		}

		// 去除空白字符
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理通知消息
		cmd := m.cmd.Load().(string)
		if m.notifications.IsNotification(line, cmd) {
			m.printf("receive urc: %s", line)
			if m.urcHandler != nil {
				go m.urcHandler(parseParam(line))
			}
			continue
		}

		// 写入响应通道
		select {
		case m.responseChan <- line:
			m.printf("collect line: %s", line)
		default:
			// 通道满了，丢弃数据（避免阻塞）
			m.printf("discard line: %s", line)
		}
	}
}

// writeString 写入数据到串口
func (m *Device) writeString(data string) error {
	if m.closed.Load() {
		return fmt.Errorf("device closed")
	}

	m.printf("send command: %s", data)

	// 向串口写入数据
	n, err := m.port.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("incomplete: wrote %d of %d bytes", n, len(data))
	}

	return nil
}
