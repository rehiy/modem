package at

import (
	"bufio"
	"io"
)

// cmdLoop 负责与调制解调器的接口，序列化命令发出并等待响应
func cmdLoop(cmds chan func(), in <-chan string, out chan struct{}) {
	for {
		select {
		case cmd := <-cmds:
			cmd()
		case _, ok := <-in:
			if !ok {
				close(out)
				return
			}
		}
	}
}

// lineReader 从m获取行并重定向到out，m关闭时退出
func lineReader(m io.Reader, out chan string) {
	scanner := bufio.NewScanner(m)
	scanner.Split(scanLines)
	for scanner.Scan() {
		out <- scanner.Text()
	}
	close(out) // tell pipeline we're done - end of pipeline will close the AT.
}

// scanLines 是lineReader的自定义行扫描器，识别调制解调器响应SMS命令（如+CMGS）返回的提示
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 处理SMS提示特殊情况 - 提示处无CR
	if len(data) >= 1 && data[0] == '>' {
		i := 1
		// 可能有尾随空格，所以吞噬它...
		for ; i < len(data) && data[i] == ' '; i++ {
		}
		return i, data[0:1], nil
	}
	return bufio.ScanLines(data, atEOF)
}
