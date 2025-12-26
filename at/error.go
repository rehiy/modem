package at

import (
	"errors"
	"strings"
)

var (
	// ErrError 表示调制解调器返回通用AT错误响应
	ErrError = errors.New("ERROR")

	// ErrClosed 表示调制解调器已关闭，无法执行操作
	ErrClosed = errors.New("closed")

	// ErrDeadlineExceeded 表示调制解调器未能在规定时间内完成操作
	ErrDeadlineExceeded = errors.New("deadline exceeded")

	// ErrIndicationExists 表示前缀已注册通知
	ErrIndicationExists = errors.New("indication exists")
)

// CMEError 表示调制解调器返回的CME错误
type CMEError string

func (e CMEError) Error() string {
	return string("CME Error: " + e)
}

// CMSError 表示调制解调器返回的CMS错误
type CMSError string

func (e CMSError) Error() string {
	return string("CMS Error: " + e)
}

// ConnectError 表示拨号尝试失败
type ConnectError string

func (e ConnectError) Error() string {
	return string("Connect: " + e)
}

// newError 解析一行并创建对应内容的错误
func newError(line string) error {
	var err error
	switch {
	case strings.HasPrefix(line, "ERROR"):
		err = ErrError
	case strings.HasPrefix(line, "+CMS ERROR:"):
		err = CMSError(strings.TrimSpace(line[11:]))
	case strings.HasPrefix(line, "+CME ERROR:"):
		err = CMEError(strings.TrimSpace(line[11:]))
	}
	return err
}
