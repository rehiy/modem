package at

import (
	"bufio"
	"io"
)

// cmdLoop is responsible for the interface to the modem.
//
// It serialises the issuing of commands and awaits the responses.
// If no command is pending then any lines received are discarded.
//
// The cmdLoop terminates when the downstream closes.
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

// lineReader takes lines from m and redirects them to out.
//
// lineReader exits when m closes.
func lineReader(m io.Reader, out chan string) {
	scanner := bufio.NewScanner(m)
	scanner.Split(scanLines)
	for scanner.Scan() {
		out <- scanner.Text()
	}
	close(out) // tell pipeline we're done - end of pipeline will close the AT.
}

// scanLines is a custom line scanner for lineReader that recognises the prompt
// returned by the modem in response to SMS commands such as +CMGS.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// handle SMS prompt special case - no CR at prompt
	if len(data) >= 1 && data[0] == '>' {
		i := 1
		// there may be trailing space, so swallow that...
		for ; i < len(data) && data[i] == ' '; i++ {
		}
		return i, data[0:1], nil
	}
	return bufio.ScanLines(data, atEOF)
}
