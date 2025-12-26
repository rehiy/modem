package at

import (
	"strings"
	"time"
)

// response 表示在调制解调器上执行的请求操作结果
type response struct {
	info []string // 命令和状态行之间返回的行集合
	err  error    // 调制解调器返回或交互时的错误
}

// Command 向调制解调器发送命令并返回结果
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

// SMSCommand 向调制解调器发送SMS命令并返回结果
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

// processReq 执行请求 - 发送命令并等待响应
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

// processSmsReq 执行SMS请求 - 发送命令，等待提示，发送数据并等待响应
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
			a.escapeWrite()
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

// processRxLine 解析从调制解调器接收的行并确定如何添加到当前命令的响应中
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

// processSmsRxLine 解析从调制解调器接收的行并确定如何添加到当前SMS命令的响应中
func (a *AT) processSmsRxLine(lt rxl, line string, sms string) (info *string, done bool, err error) {
	switch lt {
	case rxlUnknown:
		if strings.HasSuffix(line, "\x1a") && strings.HasPrefix(line, sms) {
			// swallow echoed SMS PDU
			return
		}
		info = &line
	case rxlSMSPrompt:
		if err = a.writeSMSContent(sms); err != nil {
			// escape SMS
			a.escapeWrite()
		}
	default:
		return a.processRxLine(lt, line)
	}
	return
}

// ---------------------------------- 辅助函数 ----------------------------------//

// 接收行类型
type rxl int

const (
	rxlUnknown      rxl = iota // 未知行
	rxlEchoCmdLine             // 回显命令行
	rxlInfo                    // 信息行
	rxlStatusOK                // 状态OK
	rxlStatusError             // 状态错误
	rxlAsync                   // 异步行
	rxlSMSPrompt               // SMS提示
	rxlConnect                 // 连接
	rxlConnectError            // 连接错误
)

// parseCmdID 返回命令的标识符组件，即任何'='或'?'之前的部分
func parseCmdID(cmdLine string) string {
	if idx := strings.IndexAny(cmdLine, "=?"); idx != -1 {
		return cmdLine[0:idx]
	}
	return cmdLine
}

// parseRxLine 解析接收行并识别行类型
func parseRxLine(line string, cmdID string) rxl {
	switch {
	case line == "OK":
		return rxlStatusOK
	case strings.HasPrefix(line, "ERROR"), strings.HasPrefix(line, "+CME ERROR:"), strings.HasPrefix(line, "+CMS ERROR:"):
		return rxlStatusError
	case strings.HasPrefix(line, cmdID+":"):
		return rxlInfo
	case line == ">":
		return rxlSMSPrompt
	case strings.HasPrefix(line, "AT"+cmdID):
		return rxlEchoCmdLine
	case len(cmdID) == 0 || cmdID[0] != 'D':
		// 短路非ATD命令，不在此级别识别SMS PDU
		return rxlUnknown
	case strings.HasPrefix(line, "CONNECT"):
		return rxlConnect
	case line == "BUSY", line == "NO ANSWER", line == "NO CARRIER", line == "NO DIALTONE":
		return rxlConnectError
	default:
		// 不在此级别识别SMS PDU，与其他未识别行一起捕获
		return rxlUnknown
	}
}
