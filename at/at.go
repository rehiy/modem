package at

import (
	"bufio"
	"io"
	"time"
)

// AT 表示可使用AT命令管理的调制解调器
type AT struct {
	// 发送给调制解调器的命令通道，由cmdLoop处理
	cmdCh chan func()

	// 通知变更的通道，由indLoop处理
	indCh chan func()

	// 调制解调器关闭时关闭的通道
	closed chan struct{}

	// 从调制解调器读取的所有行的通道，由indLoop处理
	iLines chan string

	// 移除通知后从调制解调器读取的行通道，由cmdLoop处理
	cLines chan string

	// 底层调制解调器，仅从cmdLoop访问
	modem io.ReadWriter

	// 转义命令和后续命令之间的最小时间间隔
	escTime time.Duration

	// 等待单个命令完成的时间
	cmdTimeout time.Duration

	// 按前缀映射的通知，仅从indLoop访问
	inds map[string]Indication

	// Init发出的命令
	initCmds []string

	// 后续命令发出前必须过期的计时器，仅从cmdLoop访问
	escGuard *time.Timer
}

// New 创建新的AT调制解调器
func New(modem io.ReadWriter, options ...Option) *AT {
	a := &AT{
		modem:      modem,
		cmdCh:      make(chan func(), 10),
		indCh:      make(chan func(), 10),
		iLines:     make(chan string, 50),
		cLines:     make(chan string, 50),
		closed:     make(chan struct{}),
		escTime:    20 * time.Millisecond,
		cmdTimeout: time.Second,
		inds:       make(map[string]Indication),
	}

	// 应用选项
	for _, option := range options {
		option.applyOption(a)
	}

	// 设置默认初始化命令
	if a.initCmds == nil {
		a.initCmds = []string{}
	}

	// 启动管道
	go lineReader(a.modem, a.iLines)
	go cmdLoop(a.cmdCh, a.cLines, a.closed)
	go a.indLoop(a.indCh, a.iLines, a.cLines)

	return a
}

// Init 执行初始化命令配置调制解调器
func (a *AT) Init(options ...InitOption) error {
	cfg := initConfig{
		cmds:    a.initCmds,
		cmdOpts: []CommandOption{},
	}

	// 应用初始化选项
	for _, option := range options {
		option.applyInitOption(&cfg)
	}

	// 执行每个初始化命令
	for _, cmd := range cfg.cmds {
		_, err := a.Command(cmd, cfg.cmdOpts...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Closed 返回一个在调制解调器未关闭时阻塞的通道
func (a *AT) Closed() <-chan struct{} {
	return a.closed
}

//---------------------------------- 辅助函数 ----------------------------------//

// cmdLoop 负责与调制解调器的接口，序列化命令发出并等待响应
func cmdLoop(cmds chan func(), in <-chan string, out chan struct{}) {
	for {
		select {
		case cmd := <-cmds:
			cmd()
		case _, ok := <-in:
			if !ok {
				close(out)
				return
			}
		}
	}
}

// lineReader 从m获取行并重定向到out，m关闭时退出
func lineReader(m io.Reader, out chan string) {
	scanner := bufio.NewScanner(m)
	scanner.Split(scanLines)
	for scanner.Scan() {
		out <- scanner.Text()
	}
	// 读取结束，关闭out通道
	close(out)
}

// scanLines 是lineReader的自定义行扫描器，识别调制解调器响应SMS命令（如+CMGS）返回的提示
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 处理SMS提示特殊情况 - 提示处无CR
	if len(data) >= 1 && data[0] == '>' {
		i := 1
		// 清理可能的尾随空格
		for ; i < len(data) && data[i] == ' '; i++ {
		}
		return i, data[0:1], nil
	}
	return bufio.ScanLines(data, atEOF)
}
