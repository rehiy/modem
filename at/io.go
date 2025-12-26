package at

import (
	"time"
)

const (
	sub = "\x1a"
	esc = "\x1b"
)

// waitEscGuard waits for a write guard to allow a write to the modem.
//
// This should only be called from within the cmdLoop.
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

// writeCommand writes a one line command to the modem.
//
// This should only be called from within the cmdLoop.
func (a *AT) writeCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r\n"
	_, err := a.modem.Write([]byte(cmdLine))
	return err
}

// writeSMSCommand writes a the first line of an SMS command to the modem.
//
// This should only be called from within the cmdLoop.
func (a *AT) writeSMSCommand(cmd string) error {
	cmdLine := "AT" + cmd + "\r"
	_, err := a.modem.Write([]byte(cmdLine))
	return err
}

// writeSMS writes the first line of a two line SMS command to the modem.
//
// This should only be called from within the cmdLoop.
func (a *AT) writeSMS(sms string) error {
	_, err := a.modem.Write([]byte(sms + string(sub)))
	return err
}

// escape issues an escape command
//
// This should only be called from within the cmdLoop.
func (a *AT) escape(b ...byte) {
	cmd := append([]byte(esc+"\r\n"), b...)
	a.modem.Write(cmd)
	a.escGuard = time.NewTimer(a.escTime)
}