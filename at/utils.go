package at

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var Terminators = []string{
	"\r",   // 回车符 (CR)
	"\n",   // 换行符 (LF)
	"\r\n", // 标准结束符 (CRLF)
	"\x1A", // Ctrl+Z (短信发送确认)
	"\x1B", // ESC (取消输入)
}

var labelRegex = regexp.MustCompile(`\+[A-Z0-9]+`)

// parseInt 解析整数
func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// hasTerminator 检查命令是否包含任何结束符
func hasTerminator(cmd string) bool {
	for _, t := range Terminators {
		if strings.HasSuffix(cmd, t) {
			return true
		}
	}
	return false
}

// parseParam 解析响应内容
func parseParam(line string) (string, map[int]string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		param := map[int]string{}
		label := strings.TrimSpace(parts[0])
		group := strings.Split(strings.TrimSpace(parts[1]), ",")
		for i, v := range group {
			param[i] = strings.Trim(strings.TrimSpace(v), `"'`)
		}
		return label, param
	}
	return line, nil
}

// getCommandResponseLabel 从 AT 命令中提取响应标签
// 例如: "AT+CLCC" -> "+CLCC", "ATD" -> "" (ATD 不带前缀，返回空)
func getCommandResponseLabel(cmd string) string {
	if label := labelRegex.FindString(cmd); label != "" {
		return label
	}
	return ""
}

// parseResponse 解析命令响应，返回第一个匹配的参数
func parseResponse(cmd string, responses []string, plen int) (map[int]string, error) {
	label := getCommandResponseLabel(cmd)
	for _, line := range responses {
		respLabel, param := parseParam(line)
		if respLabel == label && len(param) >= plen {
			return param, nil
		}
	}
	return nil, fmt.Errorf("no response matching %q found", label)
}

// parseResponseFiltered 解析命令响应，返回第一个匹配的参数（支持过滤）
func parseResponseFiltered(cmd string, responses []string, plen int, filter func(map[int]string) bool) (map[int]string, error) {
	label := getCommandResponseLabel(cmd)
	for _, line := range responses {
		respLabel, param := parseParam(line)
		if respLabel == label && len(param) >= plen && filter(param) {
			return param, nil
		}
	}
	return nil, fmt.Errorf("no response matching %q found", label)
}
