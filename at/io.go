package at

import (
	"time"
)

// writeCommand 向调制解调器写入单行命令
func (a *AT) writeCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r\n"
	_, err := a.modem.Write([]byte(cmdLine))
	return err
}

// writeSMSCommand 向调制解调器写入SMS命令的第一行
func (a *AT) writeSMSCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r"
	_, err := a.modem.Write([]byte(cmdLine))
	return err
}

// writeSMSContent 向调制解调器写入两行SMS命令的第二行
func (a *AT) writeSMSContent(sms string) error {
	_, err := a.modem.Write([]byte(sms + string("\x1a")))
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
