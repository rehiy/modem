package at

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
