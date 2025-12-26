package at

import (
	"time"
)

// Option 是AT的构造选项
type Option interface {
	applyOption(*AT)
}

// CommandOption 定义Command和SMSCommand的行为选项
type CommandOption interface {
	applyCommandOption(*commandConfig)
}

type commandConfig struct {
	timeout time.Duration // 命令超时时间
}

// InitOption 定义Init的行为选项
type InitOption interface {
	applyInitOption(*initConfig)
}

type initConfig struct {
	cmds    []string      // 初始化命令列表
	cmdOpts []CommandOption // 命令选项列表
}

// WithEscTime 设置调制解调器的守卫时间
func WithEscTime(d time.Duration) EscTimeOption {
	return EscTimeOption(d)
}

// EscTimeOption 定义调制解调器的转义守卫时间
type EscTimeOption time.Duration

func (o EscTimeOption) applyOption(a *AT) {
	a.escTime = time.Duration(o)
}

// WithCmds 指定Init发出的AT命令集
func WithCmds(cmds ...string) CmdsOption {
	return CmdsOption(cmds)
}

// CmdsOption 指定Init发出的AT命令集
type CmdsOption []string

func (o CmdsOption) applyOption(a *AT) {
	a.initCmds = []string(o)
}

func (o CmdsOption) applyInitOption(i *initConfig) {
	i.cmds = []string(o)
}

// WithTimeout 指定调制解调器完成命令的最大允许时间
func WithTimeout(d time.Duration) TimeoutOption {
	return TimeoutOption(d)
}

// TimeoutOption 指定调制解调器完成命令的最大允许时间
type TimeoutOption time.Duration

func (o TimeoutOption) applyOption(a *AT) {
	a.cmdTimeout = time.Duration(o)
}

func (o TimeoutOption) applyCommandOption(c *commandConfig) {
	c.timeout = time.Duration(o)
}

func (o TimeoutOption) applyInitOption(i *initConfig) {
	i.cmdOpts = append(i.cmdOpts, o)
}
