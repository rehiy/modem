package at

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

// commandTestPort 将模拟响应送入命令通道，测试无需依赖真实串口。
type commandTestPort struct {
	write func(string) error
}

func (p *commandTestPort) Read([]byte) (int, error) { return 0, io.EOF }
func (p *commandTestPort) Flush() error             { return nil }
func (p *commandTestPort) Close() error             { return nil }
func (p *commandTestPort) Write(data []byte) (int, error) {
	if err := p.write(string(data)); err != nil {
		return 0, err
	}
	return len(data), nil
}

func newCommandTestDevice(timeout time.Duration, respond func(string, chan<- string) error) *Device {
	responses := make(chan string, 10)
	return &Device{
		port: &commandTestPort{write: func(command string) error {
			return respond(command, responses)
		}},
		timeout:      timeout,
		commands:     *DefaultCommandSet(),
		responses:    *DefaultResponseSet(),
		responseChan: responses,
		printf:       func(string, ...any) {},
	}
}

func TestSendCommandWithTimeoutPreservesDefault(t *testing.T) {
	device := newCommandTestDevice(2*time.Second, func(command string, responses chan<- string) error {
		if command == "AT+SLOW\r\n" {
			responses <- "+SLOW: waiting"
		} else {
			// 后续命令需等待超过上一条命令的超时，验证默认超时没有被改写。
			time.AfterFunc(30*time.Millisecond, func() { responses <- "OK" })
		}
		return nil
	})

	started := time.Now()
	responses, err := device.SendCommandWithTimeout("AT+SLOW", 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "command timeout") {
		t.Fatalf("expected command timeout, got responses=%v, err=%v", responses, err)
	}
	if elapsed := time.Since(started); elapsed >= time.Second {
		t.Fatalf("per-command timeout was ignored: elapsed=%v", elapsed)
	}
	if !reflect.DeepEqual(responses, []string{"+SLOW: waiting"}) {
		t.Fatalf("partial responses were lost: %v", responses)
	}
	if responses, err = device.SendCommand("AT"); err != nil || !reflect.DeepEqual(responses, []string{"OK"}) {
		t.Fatalf("default timeout was changed: responses=%v, err=%v", responses, err)
	}
}

func TestSendCommandWithTimeoutAllowsLongResponse(t *testing.T) {
	device := newCommandTestDevice(time.Millisecond, func(command string, responses chan<- string) error {
		time.AfterFunc(30*time.Millisecond, func() { responses <- "OK" })
		return nil
	})
	responses, err := device.SendCommandWithTimeout("AT", time.Second)
	if err != nil || !reflect.DeepEqual(responses, []string{"OK"}) {
		t.Fatalf("longer timeout was not used: responses=%v, err=%v", responses, err)
	}
}

func TestSendCommandWithTimeoutFallsBackToDefault(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Second} {
		t.Run(timeout.String(), func(t *testing.T) {
			device := newCommandTestDevice(time.Second, func(command string, responses chan<- string) error {
				time.AfterFunc(10*time.Millisecond, func() { responses <- "OK" })
				return nil
			})
			if _, err := device.SendCommandWithTimeout("AT", timeout); err != nil {
				t.Fatalf("default timeout was not used: %v", err)
			}
		})
	}
}

func TestSendCommandWithTimeoutPropagatesWriteError(t *testing.T) {
	device := newCommandTestDevice(time.Second, func(string, chan<- string) error {
		return io.ErrClosedPipe
	})
	if _, err := device.SendCommandWithTimeout("AT", time.Second); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("write error was not propagated: %v", err)
	}
}
