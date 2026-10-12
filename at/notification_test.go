package at

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

type notificationTestPort struct{ net.Conn }

func (*notificationTestPort) Flush() error { return nil }

func TestCustomNotify(t *testing.T) {
	notifications := DefaultNotificationSet()
	notifications.CustomNotify = []string{"+QPING", "+MPING", ""}

	tests := []struct {
		name    string
		line    string
		command string
		want    bool
	}{
		{"idle qping", `+QPING: 0,"8.8.8.8",32,161,255`, "", true},
		{"idle mping", `+MPING: "statistics",4,0,60,75,67`, "", true},
		{"qping response", `+QPING: 0,"8.8.8.8",32,161,255`, "AT+QPING=1\r\n", false},
		{"mping response", `+MPING: "statistics",4,0,60,75,67`, "AT+MPING=1\r\n", false},
		{"unrelated command", `+QPING: 0,"8.8.8.8",32,161,255`, "AT+CSQ\r\n", true},
		{"unregistered prefix", "+GTPING: 1", "", false},
		{"empty prefix", "unrelated data", "", false},
		{"standard notification", `+CMTI: "ME",1`, "", true},
		{"command error", "+CME ERROR: 10", "AT+CPIN?\r\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notifications.IsNotification(tt.line, tt.command); got != tt.want {
				t.Fatalf("IsNotification(%q, %q)=%v, want %v", tt.line, tt.command, got, tt.want)
			}
		})
	}
}

func TestDefaultNotificationsDoNotIncludeCustomPrefixes(t *testing.T) {
	if DefaultNotificationSet().IsNotification("+QPING: 1", "") {
		t.Fatal("unconfigured custom prefix was treated as a notification")
	}
}

func TestCustomNotifyDispatchesAfterCommand(t *testing.T) {
	port, modem := net.Pipe()
	defer modem.Close()
	if err := modem.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}

	notifications := DefaultNotificationSet()
	notifications.CustomNotify = []string{"+QPING"}
	received := make(chan string, 1)
	device := New(&notificationTestPort{port}, func(label string, _ map[int]string) {
		received <- label
	}, &Config{
		Timeout:         time.Second,
		NotificationSet: notifications,
		Printf:          func(string, ...any) {},
	})
	defer device.Close()

	// 同名前缀先作为命令响应返回，命令结束后的通知再交给回调。
	go func() {
		command, err := bufio.NewReader(modem).ReadString('\n')
		if err != nil {
			t.Errorf("read command: %v", err)
			return
		}
		if command != "AT+QPING=1\r\n" {
			t.Errorf("unexpected command: %q", command)
		}
		if _, err := fmt.Fprint(modem, "+QPING: 0\r\nOK\r\n"); err != nil {
			t.Errorf("write response: %v", err)
		}
	}()

	responses, err := device.SendCommand("AT+QPING=1")
	if err != nil || !reflect.DeepEqual(responses, []string{"+QPING: 0", "OK"}) {
		t.Fatalf("custom prefix was not collected as a response: responses=%v, err=%v", responses, err)
	}
	if _, err := fmt.Fprint(modem, "+QPING: 1\r\n"); err != nil {
		t.Fatal(err)
	}
	select {
	case label := <-received:
		if label != "+QPING" {
			t.Fatalf("unexpected notification: %q", label)
		}
	case <-time.After(time.Second):
		t.Fatal("custom notification was not dispatched after the command")
	}
}
