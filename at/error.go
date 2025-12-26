package at

import (
	"errors"
	"strings"
)

var (
	// ErrClosed indicates an operation cannot be performed as the modem has
	// been closed.
	ErrClosed = errors.New("closed")

	// ErrDeadlineExceeded indicates the modem failed to complete an operation
	// within the required time.
	ErrDeadlineExceeded = errors.New("deadline exceeded")

	// ErrError indicates the modem returned a generic AT ERROR in response to
	// an operation.
	ErrError = errors.New("ERROR")

	// ErrIndicationExists indicates there is already a indication registered
	// for a prefix.
	ErrIndicationExists = errors.New("indication exists")
)

// CMEError indicates a CME Error was returned by the modem.
//
// The value is the error value, in string form, which may be the numeric or
// textual, depending on the modem configuration.
type CMEError string

// CMSError indicates a CMS Error was returned by the modem.
//
// The value is the error value, in string form, which may be the numeric or
// textual, depending on the modem configuration.
type CMSError string

// ConnectError indicates an attempt to dial failed.
//
// The value of the error is the failure indication returned by the modem.
type ConnectError string

func (e CMEError) Error() string {
	return string("CME Error: " + e)
}

func (e CMSError) Error() string {
	return string("CMS Error: " + e)
}

func (e ConnectError) Error() string {
	return string("Connect: " + e)
}

// newError parses a line and creates an error corresponding to the content.
func newError(line string) error {
	var err error
	switch {
	case strings.HasPrefix(line, "ERROR"):
		err = ErrError
	case strings.HasPrefix(line, "+CMS ERROR:"):
		err = CMSError(strings.TrimSpace(line[11:]))
	case strings.HasPrefix(line, "+CME ERROR:"):
		err = CMEError(strings.TrimSpace(line[11:]))
	}
	return err
}
