package at

import (
	"strings"
)

// Received line types.
type rxl int

const (
	rxlUnknown rxl = iota
	rxlEchoCmdLine
	rxlInfo
	rxlStatusOK
	rxlStatusError
	rxlAsync
	rxlSMSPrompt
	rxlConnect
	rxlConnectError
)

// parseCmdID returns the identifier component of the command.
//
// This is the section prior to any '=' or '?' and is generally, but not
// always, used to prefix info lines corresponding to the command.
func parseCmdID(cmdLine string) string {
	if idx := strings.IndexAny(cmdLine, "=?"); idx != -1 {
		return cmdLine[0:idx]
	}
	return cmdLine
}

// parseRxLine parses a received line and identifies the line type.
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
		// Short circuit non-ATD commands.
		// No attempt to identify SMS PDUs at this level, so they will
		// be caught here, along with other unidentified lines.
		return rxlUnknown
	case strings.HasPrefix(line, "CONNECT"):
		return rxlConnect
	case line == "BUSY", line == "NO ANSWER", line == "NO CARRIER", line == "NO DIALTONE":
		return rxlConnectError
	default:
		// No attempt to identify SMS PDUs at this level, so they will
		// be caught here, along with other unidentified lines.
		return rxlUnknown
	}
}
