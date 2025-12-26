package at

import (
	"time"
)

// Option is a construction option for an AT.
type Option interface {
	applyOption(*AT)
}

// CommandOption defines a behaviouralk option for Command and SMSCommand.
type CommandOption interface {
	applyCommandOption(*commandConfig)
}

// InitOption defines a behaviouralk option for Init.
type InitOption interface {
	applyInitOption(*initConfig)
}

// WithEscTime sets the guard time for the modem.
//
// The escape time is the minimum time between an escape command being sent to
// the modem and any subsequent commands.
//
// The default guard time is 20msec.
func WithEscTime(d time.Duration) EscTimeOption {
	return EscTimeOption(d)
}

// EscTimeOption defines the escape guard time for the modem.
type EscTimeOption time.Duration

func (o EscTimeOption) applyOption(a *AT) {
	a.escTime = time.Duration(o)
}

// CmdsOption specifies the set of AT commands issued by Init.
type CmdsOption []string

func (o CmdsOption) applyOption(a *AT) {
	a.initCmds = []string(o)
}

func (o CmdsOption) applyInitOption(i *initConfig) {
	i.cmds = []string(o)
}

// WithCmds specifies the set of AT commands issued by Init.
//
// The default commands are ATZ.
func WithCmds(cmds ...string) CmdsOption {
	return CmdsOption(cmds)
}

// WithTimeout specifies the maximum time allowed for the modem to complete a
// command.
func WithTimeout(d time.Duration) TimeoutOption {
	return TimeoutOption(d)
}

// TimeoutOption specifies the maximum time allowed for the modem to complete a
// command.
type TimeoutOption time.Duration

func (o TimeoutOption) applyOption(a *AT) {
	a.cmdTimeout = time.Duration(o)
}

func (o TimeoutOption) applyInitOption(i *initConfig) {
	i.cmdOpts = append(i.cmdOpts, o)
}

func (o TimeoutOption) applyCommandOption(c *commandConfig) {
	c.timeout = time.Duration(o)
}

type commandConfig struct {
	timeout time.Duration
}

type initConfig struct {
	cmds    []string
	cmdOpts []CommandOption
}
