package at

import (
	"time"
)

const (
	sub = "\x1a"
	esc = "\x1b"
)

// waitEscGuard 等待写守卫允许向调制解调器写入，应仅在cmdLoop中调用
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

// writeCommand 向调制解调器写入单行命令，应仅在cmdLoop中调用
func (a *AT) writeCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r\n"
	_, err := a.modem.Write([]byte(cmdLine))
	return err
}

// writeSMSCommand 向调制解调器写入SMS命令的第一行，应仅在cmdLoop中调用
func (a *AT) writeSMSCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r"
	_, err := a.modem.Write([]byte(cmdLine))
	return err
}

// writeSMS 向调制解调器写入两行SMS命令的第二行，应仅在cmdLoop中调用
func (a *AT) writeSMS(sms string) error {
	_, err := a.modem.Write([]byte(sms + string(sub)))
	return err
}

// escape 发出转义命令，应仅在cmdLoop中调用
func (a *AT) escape(b ...byte) {
	cmd := append([]byte(esc+"\r\n"), b...)
	a.modem.Write(cmd)
	a.escGuard = time.NewTimer(a.escTime)
}
