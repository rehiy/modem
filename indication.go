package at

import (
	"strings"
)

// InfoHandler receives indication info.
type InfoHandler func([]string)

// Indication represents an unsolicited result code (URC) from the modem, such
// as a received SMS message.
//
// Indications are lines prefixed with a particular pattern, and may include a
// number of trailing lines. The matching lines are bundled into a slice and
// sent to the handler.
type Indication struct {
	prefix  string
	lines   int
	handler InfoHandler
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

// IndicationOption alters the behavior of the indication.
type IndicationOption interface {
	applyIndicationOption(*Indication)
}

// TrailingLinesOption specifies the number of trailing lines expected after an
// indication line.
type TrailingLinesOption int

func (o TrailingLinesOption) applyIndicationOption(ind *Indication) {
	ind.lines = int(o) + 1

}

// WithTrailingLines indicates the number of lines after the line containing
// the indication that arew to be collected as part of the indication.
//
// The default is 0 - only the indication line itself is collected and returned.
func WithTrailingLines(l int) TrailingLinesOption {
	return TrailingLinesOption(l)
}

// WithTrailingLine indicates the indication includes one line after the line
// containing the indication.
var WithTrailingLine = TrailingLinesOption(1)

// WithIndication adds an indication during construction.
func WithIndication(prefix string, handler InfoHandler, options ...IndicationOption) Indication {
	return newIndication(prefix, handler, options...)
}

func (o Indication) applyOption(a *AT) {
	a.inds[o.prefix] = o
}

// AddIndication adds a handler for a set of lines beginning with the prefixed
// line and the following trailing lines.
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

// CancelIndication removes any indication corresponding to the prefix.
//
// If any such indication exists its return channel is closed and no further
// indications will be sent to it.
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

// indLoop is responsible for pulling indications from the stream of lines read
// from the modem, and forwarding them to handlers.
//
// Non-indication lines are passed upstream. Indication trailing lines are
// assumed to arrive in a contiguous block immediately after the indication.
//
// indLoop exits when the in channel closes.
func (a *AT) indLoop(cmds chan func(), in <-chan string, out chan string) {
	defer close(out)
	for {
		select {
		case cmd := <-cmds:
			cmd()
		case line, ok := <-in:
			if !ok {
				return
			}
			for prefix, ind := range a.inds {
				if strings.HasPrefix(line, prefix) {
					n := make([]string, ind.lines)
					n[0] = line
					for i := 1; i < ind.lines; i++ {
						t, ok := <-in
						if !ok {
							return
						}
						n[i] = t
					}
					go ind.handler(n)
					continue
				}
			}
			out <- line
		}
	}
}
