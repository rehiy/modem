package at

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
