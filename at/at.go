package at

import (
	"io"
	"time"
)

// AT 表示可使用AT命令管理的调制解调器
type AT struct {
	// 底层调制解调器，仅从cmdLoop访问
	modem io.ReadWriter

	// 发送给调制解调器的命令通道，由cmdLoop处理
	cmdCh chan func()

	// 通知变更的通道，由indLoop处理
	indCh chan func()

	// 移除通知后从调制解调器读取的行通道，由cmdLoop处理
	cLines chan string

	// 从调制解调器读取的所有行的通道，由indLoop处理
	iLines chan string

	// 调制解调器关闭时关闭的通道
	closed chan struct{}

	// Init发出的命令
	initCmds []string

	// 按前缀映射的通知，仅从indLoop访问
	inds map[string]Indication

	// 等待单个命令完成的时间
	cmdTimeout time.Duration

	// 转义命令和后续命令之间的最小时间间隔
	escTime time.Duration

	// 后续命令发出前必须过期的计时器
	escGuard *time.Timer
}

// New 创建新的AT调制解调器
func New(modem io.ReadWriter, options ...Option) *AT {
	a := &AT{
		modem:      modem,
		cmdCh:      make(chan func(), 10),
		indCh:      make(chan func(), 10),
		cLines:     make(chan string, 50),
		iLines:     make(chan string, 50),
		closed:     make(chan struct{}),
		initCmds:   []string{},
		inds:       make(map[string]Indication),
		escTime:    20 * time.Millisecond,
		cmdTimeout: time.Second,
	}

	// 应用选项
	for _, option := range options {
		option.applyOption(a)
	}

	// 启动管道
	go a.cmdLoop()
	go a.indLoop()
	go a.lineReader()

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

// cmdLoop 负责与调制解调器的接口，序列化命令发出并等待响应
func (a *AT) cmdLoop() {
	for {
		select {
		case cmd := <-a.cmdCh:
			cmd()
		case _, ok := <-a.cLines:
			if !ok {
				close(a.closed)
				return
			}
		}
	}
}
