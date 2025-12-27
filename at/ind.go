package at

import (
	"strings"
)

// indLoop 从调制解调器读取行流并转发给处理器
func (a *AT) indLoop() {
	defer close(a.cLines)
	for {
		select {
		case cmd := <-a.indCh:
			cmd()
		case line, ok := <-a.iLines:
			if !ok {
				return
			}
			for prefix, ind := range a.inds {
				if strings.HasPrefix(line, prefix) {
					n := make([]string, ind.lines)
					n[0] = line
					for i := 1; i < ind.lines; i++ {
						t, ok := <-a.iLines
						if !ok {
							return
						}
						n[i] = t
					}
					go ind.handler(n)
					continue
				}
			}
			// 未处理的行发送到cLines
			a.cLines <- line
		}
	}
}

// AddIndication 添加指定前缀和后续行的处理器
func (a *AT) AddIndication(prefix string, handler InfoHandler, options ...IndicationOption) (err error) {
	ind := newIndication(prefix, handler, options...)
	errs := make(chan error)
	indf := func() {
		if _, ok := a.inds[ind.prefix]; ok {
			errs <- ErrIndicationExists
			return
		}
		a.inds[ind.prefix] = ind
		close(errs)
	}
	select {
	case <-a.closed:
		err = ErrClosed
	case a.indCh <- indf:
		err = <-errs
	}
	return
}

// CancelIndication 移除指定前缀的通知
func (a *AT) CancelIndication(prefix string) {
	done := make(chan struct{})
	indf := func() {
		delete(a.inds, prefix)
		close(done)
	}
	select {
	case <-a.closed:
	case a.indCh <- indf:
		<-done
	}
}

//---------------------------------- 辅助函数 ----------------------------------//

// InfoHandler 处理通知信息
type InfoHandler func([]string)

// Indication 表示来自调制解调器的非请求结果码，如接收到的SMS消息
type Indication struct {
	prefix  string
	lines   int
	handler InfoHandler
}

func (o Indication) applyOption(a *AT) {
	a.inds[o.prefix] = o
}

func newIndication(prefix string, handler InfoHandler, options ...IndicationOption) Indication {
	ind := Indication{
		prefix:  prefix,
		handler: handler,
		lines:   1,
	}
	for _, option := range options {
		option.applyIndicationOption(&ind)
	}
	return ind
}

// IndicationOption 修改通知行为
type IndicationOption interface {
	applyIndicationOption(*Indication)
}

// TrailingLinesOption 指定通知行后的后续行数
type TrailingLinesOption int

func (o TrailingLinesOption) applyIndicationOption(ind *Indication) {
	ind.lines = int(o) + 1
}

// WithTrailingLine 包含一行后续行
var WithTrailingLine = TrailingLinesOption(1)

// WithTrailingLines 设置通知行后的收集行数
func WithTrailingLines(l int) TrailingLinesOption {
	return TrailingLinesOption(l)
}

// WithIndication 构造时添加通知
func WithIndication(prefix string, handler InfoHandler, options ...IndicationOption) Indication {
	return newIndication(prefix, handler, options...)
}
