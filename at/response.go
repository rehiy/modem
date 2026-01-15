package at

import (
	"reflect"
	"strings"
)

// ResponseSet 定义可配置的命令最终响应类型集合
type ResponseSet struct {
	// 基本响应
	OK    string // OK - 成功响应
	Error string // ERROR - 通用错误响应

	// 通话/连接相关结果码
	NoCarrier  string // NO CARRIER - 无载波/连接丢失
	NoAnswer   string // NO ANSWER - 无应答
	NoDialtone string // NO DIALTONE - 无拨号音
	Busy       string // BUSY - 对方忙线
	Connect    string // CONNECT - 连接成功（可能附带速度如 "CONNECT 9600"）

	// 错误响应
	CMEError string // +CME ERROR - 移动设备错误
	CMSError string // +CMS ERROR - 短信服务错误
	CISError string // +CIS ERROR - 通信识别模块错误

	// 提示符
	Prompt string // > - 短信输入提示符

	// 自定义响应（厂商扩展）
	CustomFinal []string // 自定义最终响应列表（非标准）
}

// DefaultResponseSet 返回默认的命令响应类型集合
func DefaultResponseSet() *ResponseSet {
	return &ResponseSet{
		// 基本响应
		OK:    "OK",
		Error: "ERROR",

		// 通话/连接相关结果码
		NoCarrier:  "NO CARRIER",
		NoAnswer:   "NO ANSWER",
		NoDialtone: "NO DIALTONE",
		Busy:       "BUSY",
		Connect:    "CONNECT",

		// 错误响应
		CMEError: "+CME ERROR",
		CMSError: "+CMS ERROR",
		CISError: "+CIS ERROR",

		// 提示符
		Prompt: ">",

		// 自定义响应
		CustomFinal: []string{},
	}
}

// GetAllResponses 返回所有最终响应的列表
func (rs *ResponseSet) GetAllResponses() []string {
	v := reflect.ValueOf(rs).Elem()

	responses := []string{}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// 处理字符串类型字段（不包括 CustomFinal 切片）
		if field.Kind() == reflect.String {
			value := field.String()
			if value != "" {
				responses = append(responses, value)
			}
		}
	}

	// 添加自定义最终响应列表
	return append(responses, rs.CustomFinal...)
}

// IsFinal 检查是否为最终响应
func (rs *ResponseSet) IsFinal(line string) bool {
	for _, resp := range rs.GetAllResponses() {
		if resp != "" && strings.HasPrefix(line, resp) {
			return true
		}
	}
	return false
}

// IsError 检查是否为错误响应
func (rs *ResponseSet) IsError(line string) bool {
	responses := []string{
		rs.Error,
		rs.NoCarrier,
		rs.NoAnswer,
		rs.NoDialtone,
		rs.Busy,
		rs.CMEError,
		rs.CMSError,
		rs.CISError,
	}
	for _, resp := range responses {
		if resp != "" && strings.HasPrefix(line, resp) {
			return true
		}
	}
	return false
}

// IsSuccess 检查是否为成功响应
func (rs *ResponseSet) IsSuccess(line string) bool {
	responses := []string{
		rs.OK,
		rs.Connect,
		rs.Prompt,
	}
	for _, resp := range responses {
		if resp != "" && strings.HasPrefix(line, resp) {
			return true
		}
	}
	return false
}
