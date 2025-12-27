package at

import (
	"bufio"
	"time"

	"github.com/sirupsen/logrus"
)

func (a *AT) writeString(cmd string) (int, error) {
	logrus.Debugf("<< TX: %s", cmd)
	return a.modem.Write([]byte(cmd))
}

// writeCommand 向调制解调器写入单行命令
func (a *AT) writeCommand(cmd string) error {
	_, err := a.writeString("AT" + cmd + "\r\n")
	return err
}

// writeSMSCommand 向调制解调器写入SMS命令的第一行
func (a *AT) writeSMSCommand(cmd string) error {
	_, err := a.writeString("AT" + cmd + "\r")
	return err
}

// writeSMSContent 向调制解调器写入两行SMS命令的第二行
func (a *AT) writeSMSContent(sms string) error {
	_, err := a.writeString(sms + "\x1a")
	return err
}

// writeEscape 结束写守卫并向调制解调器写入数据
func (a *AT) writeEscape() error {
	_, err := a.writeString("\x1b\r\n")
	a.escGuard = time.NewTimer(a.escTime)
	return err
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

// readLoop 从调制解调器获取行并重定向到iLines，调制解调器关闭时退出
func (a *AT) readLoop() {
	scanner := bufio.NewScanner(a.modem)
	//  自定义行扫描器，识别调制解调器响应SMS命令（如+CMGS）返回的提示
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// 处理SMS提示特殊情况 - 提示处无CR
		if len(data) >= 1 && data[0] == '>' {
			i := 1
			for ; i < len(data) && data[i] == ' '; i++ {
				// 清理可能的尾随空格
			}
			return i, data[0:1], nil
		}
		return bufio.ScanLines(data, atEOF)
	})
	// 读取行并发送到iLines通道
	for scanner.Scan() {
		line := scanner.Text()
		logrus.Debugf(">> RX: %d bytes", len(line))
		a.iLines <- line
	}
	// 检查扫描错误
	if err := scanner.Err(); err != nil {
		logrus.Debugf("<> ERR: %v", err)
	}
	// 关闭iLines通道
	close(a.iLines)
}
