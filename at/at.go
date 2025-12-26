package at

import (
	"io"
	"time"
)

// AT represents a modem that can be managed using AT commands.
//
// Commands can be issued to the modem using the Command and SMSCommand methods.
//
// The AT closes the closed channel when the connection to the underlying
// modem is broken (Read returns EOF).
//
// When closed, all outstanding commands return ErrClosed and the state of the
// underlying modem becomes unknown.
//
// Once closed the AT cannot be re-opened - it must be recreated.
type AT struct {
	// channel for commands issued to the modem
	//
	// Handled by the cmdLoop.
	cmdCh chan func()

	// channel for changes to inds
	//
	// Handled by the indLoop.
	indCh chan func()

	// closed when modem is closed
	closed chan struct{}

	// channel for all lines read from the modem
	//
	// Handled by the indLoop.
	iLines chan string

	// channel for lines read from the modem after indications removed
	//
	// Handled by the cmdLoop.
	cLines chan string

	// the underlying modem
	//
	// Only accessed from the cmdLoop.
	modem io.ReadWriter

	// the minimum time between an escape command and the subsequent command
	escTime time.Duration

	// time to wait for individual commands to complete
	cmdTimeout time.Duration

	// indications mapped by prefix
	//
	// Only accessed from the indLoop
	inds map[string]Indication

	// commands issued by Init.
	initCmds []string

	// if not-nil, the timer that must expire before the subsequent command is issued
	//
	// Only accessed from the cmdLoop.
	escGuard *time.Timer
}

// New creates a new AT modem.
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

	// Apply Options
	for _, option := range options {
		option.applyOption(a)
	}

	// Set default initCmds
	if a.initCmds == nil {
		a.initCmds = []string{}
	}

	// Start the pipeline
	go lineReader(a.modem, a.iLines)
	go a.indLoop(a.indCh, a.iLines, a.cLines)
	go cmdLoop(a.cmdCh, a.cLines, a.closed)

	return a
}

// Init executes initialization commands to configure the modem.
//
// This method sends the initialization command sequence to the modem
// and verifies that the modem is responding correctly.
//
// Returns an error if any initialization command fails or if the modem
// does not respond within the configured timeout.
func (a *AT) Init(options ...InitOption) error {
	cfg := initConfig{
		cmds:    a.initCmds,
		cmdOpts: []CommandOption{},
	}

	// Apply InitOptions
	for _, option := range options {
		option.applyInitOption(&cfg)
	}

	// Execute each initialization command
	for _, cmd := range cfg.cmds {
		_, err := a.Command(cmd, cfg.cmdOpts...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Closed returns a channel which will block while the modem is not closed.
func (a *AT) Closed() <-chan struct{} {
	return a.closed
}
