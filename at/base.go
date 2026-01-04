package at

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 串口
type Port interface {
	Read(buf []byte) (int, error)
	Write(data []byte) (int, error)
	Flush() error
	Close() error
}

// 结束符
var Terminators = []string{
	"\r\n", // 标准结束符 (CRLF)
	"\n",   // 换行符 (LF)
	"\r",   // 回车符 (CR)
	"\x1A", // Ctrl+Z (短信发送确认)
	"\x1B", // ESC (取消输入)
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
	name          string               // 设备名称
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
	hasTerminator := false
	for _, item := range Terminators {
		if strings.HasSuffix(cmd, item) {
			hasTerminator = true
			break
		}
	}
	if !hasTerminator {
		cmd = cmd + Terminators[0]
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
				return responses, fmt.Errorf("device closed")
			}

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

		m.printf("read line: %s", line)

		// 处理通知消息
		cmd := m.cmd.Load().(string)
		if m.notifications.IsNotification(line, cmd) {
			if m.urcHandler != nil {
				go m.urcHandler(parseParam(line))
			}
			continue
		}

		// 将数据写入响应通道
		select {
		case m.responseChan <- line:
		default:
			// 通道满了，丢弃数据（避免阻塞）
			m.printf("discarding data: %s", line)
		}
	}
}

// writeString 写入数据到串口
func (m *Device) writeString(data string) error {
	if m.closed.Load() {
		return fmt.Errorf("device closed")
	}

	m.printf("write cmd: %s", data)

	// 向串口写入数据
	n, err := m.port.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("incomplete write: wrote %d of %d bytes", n, len(data))
	}

	return nil
}

// ===== 辅助工具 =====

// parseInt 解析整数
func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// parseParam 解析响应内容
func parseParam(line string) (string, map[int]string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		label := strings.TrimSpace(parts[0])
		group := strings.Split(strings.TrimSpace(parts[1]), ",")
		param := map[int]string{}
		for i, v := range group {
			param[i] = strings.Trim(strings.TrimSpace(v), `"'`)
		}
		return label, param
	}
	return line, nil
}
