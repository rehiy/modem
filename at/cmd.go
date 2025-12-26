package at

import (
	"strings"
	"time"
)

// response represents the result of a request operation performed on the
// modem.
//
// info is the collection of lines returned between the command and the status
// line. err corresponds to any error returned by the modem or while
// interacting with the modem.
type response struct {
	info []string
	err  error
}

// Command issues the command to the modem and returns the result.
//
// The command should NOT include the AT prefix, nor <CR><LF> suffix which is
// automatically added.
//
// The return value includes the info (the lines returned by the modem between
// the command and the status line), or an error if the command did not
// complete successfully.
func (a *AT) Command(cmd string, options ...CommandOption) ([]string, error) {
	cfg := commandConfig{timeout: a.cmdTimeout}
	for _, option := range options {
		option.applyCommandOption(&cfg)
	}
	done := make(chan response)
	cmdf := func() {
		info, err := a.processReq(cmd, cfg.timeout)
		done <- response{info: info, err: err}
	}
	select {
	case <-a.closed:
		return nil, ErrClosed
	case a.cmdCh <- cmdf:
		rsp := <-done
		return rsp.info, rsp.err
	}
}

// SMSCommand issues an SMS command to the modem, and returns the result.
//
// An SMS command is issued in two steps; first the command line:
//
//	AT<command><CR>
//
// which the modem responds to with a ">" prompt, after which the SMS PDU is
// sent to the modem:
//
//	<sms><Ctrl-Z>
//
// The modem then completes the command as per other commands, such as those
// issued by Command.
//
// The format of the sms may be a text message or a hex coded SMS PDU,
// depending on the modem configuration (text or PDU mode).
func (a *AT) SMSCommand(cmd string, sms string, options ...CommandOption) (info []string, err error) {
	cfg := commandConfig{timeout: a.cmdTimeout}
	for _, option := range options {
		option.applyCommandOption(&cfg)
	}
	done := make(chan response)
	cmdf := func() {
		info, err := a.processSmsReq(cmd, sms, cfg.timeout)
		done <- response{info: info, err: err}
	}
	select {
	case <-a.closed:
		return nil, ErrClosed
	case a.cmdCh <- cmdf:
		rsp := <-done
		return rsp.info, rsp.err
	}
}

// processReq performs a request - issuing the command and awaiting the response.
func (a *AT) processReq(cmd string, timeout time.Duration) (info []string, err error) {
	a.waitEscGuard()
	err = a.writeCommand(cmd)
	if err != nil {
		return
	}

	cmdID := parseCmdID(cmd)
	var expChan <-chan time.Time
	if timeout >= 0 {
		expiry := time.NewTimer(timeout)
		expChan = expiry.C
		defer expiry.Stop()
	}
	for {
		select {
		case <-expChan:
			err = ErrDeadlineExceeded
			return
		case line, ok := <-a.cLines:
			if !ok {
				return nil, ErrClosed
			}
			if line == "" {
				continue
			}
			lt := parseRxLine(line, cmdID)
			i, done, perr := a.processRxLine(lt, line)
			if i != nil {
				info = append(info, *i)
			}
			if perr != nil {
				err = perr
				return
			}
			if done {
				return
			}
		}
	}
}

// processSmsReq performs a SMS request - issuing the command, awaiting the prompt, sending
// the data and awaiting the response.
func (a *AT) processSmsReq(cmd string, sms string, timeout time.Duration) (info []string, err error) {
	a.waitEscGuard()
	err = a.writeSMSCommand(cmd)
	if err != nil {
		return
	}
	cmdID := parseCmdID(cmd)
	var expChan <-chan time.Time
	if timeout >= 0 {
		expiry := time.NewTimer(timeout)
		expChan = expiry.C
		defer expiry.Stop()
	}
	for {
		select {
		case <-expChan:
			// cancel outstanding SMS request
			a.escape()
			err = ErrDeadlineExceeded
			return
		case line, ok := <-a.cLines:
			if !ok {
				err = ErrClosed
				return
			}
			if line == "" {
				continue
			}
			lt := parseRxLine(line, cmdID)
			i, done, perr := a.processSmsRxLine(lt, line, sms)
			if i != nil {
				info = append(info, *i)
			}
			if perr != nil {
				err = perr
				return
			}
			if done {
				return
			}
		}
	}
}

// processRxLine parses a line received from the modem and determines how it
// adds to the response for the current command.
//
// The return values are:
//   - a line of info to be added to the response (optional)
//   - a flag indicating if the command is complete.
//   - an error detected while processing the command.
func (a *AT) processRxLine(lt rxl, line string) (info *string, done bool, err error) {
	switch lt {
	case rxlStatusOK:
		done = true
	case rxlStatusError:
		err = newError(line)
	case rxlUnknown, rxlInfo:
		info = &line
	case rxlConnect:
		info = &line
		done = true
	case rxlConnectError:
		err = ConnectError(line)
	}
	return
}

// processSmsRxLine parses a line received from the modem and determines how it
// adds to the response for the current command.
//
// The return values are:
//   - a line of info to be added to the response (optional)
//   - a flag indicating if the command is complete.
//   - an error detected while processing the command.
func (a *AT) processSmsRxLine(lt rxl, line string, sms string) (info *string, done bool, err error) {
	switch lt {
	case rxlUnknown:
		if strings.HasSuffix(line, sub) && strings.HasPrefix(line, sms) {
			// swallow echoed SMS PDU
			return
		}
		info = &line
	case rxlSMSPrompt:
		if err = a.writeSMS(sms); err != nil {
			// escape SMS
			a.escape()
		}
	default:
		return a.processRxLine(lt, line)
	}
	return
}