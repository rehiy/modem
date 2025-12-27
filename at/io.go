package at

import (
	"bufio"
	"time"
)

func (a *AT) writeString(cmd string) (int, error) {
	return a.modem.Write([]byte(cmd))
}

// writeCommand 向调制解调器写入单行命令
func (a *AT) writeCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r\n"
	_, err := a.writeString(cmdLine)
	return err
}

// writeSMSCommand 向调制解调器写入SMS命令的第一行
func (a *AT) writeSMSCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r"
	_, err := a.writeString(cmdLine)
	return err
}

// writeSMSContent 向调制解调器写入两行SMS命令的第二行
func (a *AT) writeSMSContent(sms string) error {
	txtLine := sms + string("\x1a")
	_, err := a.writeString(txtLine)
	return err
}

// escapeWrite 结束写守卫并向调制解调器写入数据
func (a *AT) escapeWrite(b ...byte) {
	cmd := append([]byte("\x1b\r\n"), b...)
	a.modem.Write(cmd)
	a.escGuard = time.NewTimer(a.escTime)
}

// waitEscGuard 等待写守卫允许向调制解调器写入
func (a *AT) waitEscGuard() {
	if a.escGuard == nil {
		return
	}
	defer func() { a.escGuard = nil }()
	for {
		select {
		case _, ok := <-a.cLines:
			if !ok {
				a.escGuard.Stop()
				return
			}
		case <-a.escGuard.C:
			return
		}
	}
}

// lineReader 从调制解调器获取行并重定向到iLines，调制解调器关闭时退出
func (a *AT) lineReader() {
	scanner := bufio.NewScanner(a.modem)
	scanner.Split(scanLines)
	for scanner.Scan() {
		a.iLines <- scanner.Text()
	}
	// 读取结束，关闭iLines通道
	close(a.iLines)
}

// scanLines 自定义行扫描器，识别调制解调器响应SMS命令（如+CMGS）返回的提示
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 处理SMS提示特殊情况 - 提示处无CR
	if len(data) >= 1 && data[0] == '>' {
		i := 1
		for ; i < len(data) && data[i] == ' '; i++ {
			// 清理可能的尾随空格
		}
		return i, data[0:1], nil
	}
	return bufio.ScanLines(data, atEOF)
}
