package pdu

import (
	"testing"
	"time"
)

// TestEncodeDecode7Bit 测试 GSM 7-bit 编码和解码
func TestEncodeDecode7Bit(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"Simple", "Hello"},
		{"With space", "Hello World"},
		{"Numbers", "Test123"},
		{"Symbols", "Price: $10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode7Bit(tt.text)
			if err != nil {
				t.Fatalf("Encode7Bit failed: %v", err)
			}
			decoded := Decode7Bit(encoded, len([]rune(tt.text)))
			if decoded != tt.text {
				t.Errorf("Decode mismatch: got %q, want %q", decoded, tt.text)
			}
		})
	}
}

// TestEncodeDecode7BitExtended 测试 GSM 7-bit 扩展字符
func TestEncodeDecode7BitExtended(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"Euro", "Price: €10"},
		{"Brackets", "[test]"},
		{"Braces", "{data}"},
		{"Pipe", "a|b"},
		{"Backslash", "path\\file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode7Bit(tt.text)
			if err != nil {
				t.Fatalf("Encode7Bit failed: %v", err)
			}
			// 扩展字符占2个septets
			expectedLen := 0
			for _, r := range tt.text {
				if _, ok := gsm7bitExtChars[r]; ok {
					expectedLen += 2
				} else {
					expectedLen++
				}
			}
			decoded := Decode7Bit(encoded, expectedLen)
			if decoded != tt.text {
				t.Errorf("Decode mismatch: got %q, want %q", decoded, tt.text)
			}
		})
	}
}

// TestEncodeDecodeUCS2 测试 UCS2 编码和解码
func TestEncodeDecodeUCS2(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"Chinese", "你好世界"},
		{"Japanese", "こんにちは"},
		{"Emoji", "Hello 😀"},
		{"Mixed", "Hello世界"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeUCS2(tt.text)
			decoded := DecodeUCS2(encoded)
			if decoded != tt.text {
				t.Errorf("Decode mismatch: got %q, want %q", decoded, tt.text)
			}
		})
	}
}

// TestEncodeDecodePhoneNumber 测试电话号码编码和解码
func TestEncodeDecodePhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected string
	}{
		{"International", "+8613800138000", "+8613800138000"},
		{"Local", "13800138000", "13800138000"},
		{"With spaces", "+86 138 0013 8000", "+8613800138000"},
		{"Odd length", "+861380013800", "+861380013800"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrType, encoded := EncodePhoneNumber(tt.number)
			decoded := DecodePhoneNumber(encoded, addrType)
			if decoded != tt.expected {
				t.Errorf("Decode mismatch: got %q, want %q", decoded, tt.expected)
			}
		})
	}
}

// TestEncodeSingleSMS 测试单条短信编码
func TestEncodeSingleSMS(t *testing.T) {
	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        "Hello World",
		SMSC:        "+8613800138000",
	}

	pdus, err := Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(pdus) != 1 {
		t.Errorf("Expected 1 PDU, got %d", len(pdus))
	}

	if pdus[0].Data == "" {
		t.Error("PDU data is empty")
	}

	if pdus[0].Length == 0 {
		t.Error("PDU length is 0")
	}
}

// TestEncodeLongSMS 测试长短信编码
func TestEncodeLongSMS(t *testing.T) {
	longText := ""
	for i := 0; i < 200; i++ {
		longText += "a"
	}

	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        longText,
		SMSC:        "+8613800138000",
	}

	pdus, err := Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(pdus) != 2 {
		t.Errorf("Expected 2 PDUs, got %d", len(pdus))
	}

	for i, pdu := range pdus {
		if pdu.Data == "" {
			t.Errorf("PDU %d data is empty", i)
		}
	}
}

// TestEncodeUCS2LongSMS 测试 UCS2 长短信编码
func TestEncodeUCS2LongSMS(t *testing.T) {
	longText := ""
	for i := 0; i < 100; i++ {
		longText += "你"
	}

	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        longText,
		SMSC:        "+8613800138000",
		Encoding:    EncodingUCS2,
	}

	pdus, err := Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(pdus) < 2 {
		t.Errorf("Expected at least 2 PDUs, got %d", len(pdus))
	}
}

// TestDecodeSMSDeliver 测试解码接收的短信
func TestDecodeSMSDeliver(t *testing.T) {
	// 真实的 SMS-DELIVER PDU
	pduStr := "07911326040000F0040B911346610089F60000208062917314080CC8329BFD06"

	msg, err := Decode(pduStr)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != MessageTypeSMSDeliver {
		t.Errorf("Expected SMS-DELIVER, got %d", msg.Type)
	}

	if msg.Text == "" {
		t.Error("Decoded text is empty")
	}
}

// TestDecodeSMSSubmit 测试解码发送的短信
func TestDecodeSMSSubmit(t *testing.T) {
	// SMS-SUBMIT PDU
	pduStr := "0011000D91683108108300F00008A70C4F60597DFF0C4E16754CFF01"

	msg, err := Decode(pduStr)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if msg.Type != MessageTypeSMSSubmit {
		t.Errorf("Expected SMS-SUBMIT, got %d", msg.Type)
	}
}

// TestEncodeDecodeRoundTrip 测试编码解码往返
func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		encoding Encoding
	}{
		{"UCS2 English", "Test message", EncodingUCS2},
		{"UCS2 Chinese", "你好世界", EncodingUCS2},
		{"UCS2 with symbols", "Price: €10", EncodingUCS2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &Message{
				PhoneNumber: "+8613800138000",
				Text:        tt.text,
				SMSC:        "+8613800138000",
				Encoding:    tt.encoding,
			}

			pdus, err := Encode(original)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := Decode(pdus[0].Data)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if decoded.PhoneNumber != original.PhoneNumber {
				t.Errorf("Phone number mismatch: got %q, want %q", decoded.PhoneNumber, original.PhoneNumber)
			}

			if decoded.Text != original.Text {
				t.Errorf("Text mismatch: got %q, want %q", decoded.Text, original.Text)
			}
		})
	}
}

// Test7BitRoundTrip 测试 7-bit 编码往返
func Test7BitRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"7-bit English", "Test message"},
		{"7-bit with numbers", "Code: 12345"},
		{"7-bit with symbols", "Price: $10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &Message{
				PhoneNumber: "+8613800138000",
				Text:        tt.text,
				SMSC:        "+8613800138000",
				Encoding:    Encoding7Bit,
			}

			pdus, err := Encode(original)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := Decode(pdus[0].Data)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if decoded.PhoneNumber != original.PhoneNumber {
				t.Errorf("Phone number mismatch: got %q, want %q", decoded.PhoneNumber, original.PhoneNumber)
			}

			if decoded.Text != original.Text {
				t.Errorf("Text mismatch: got %q, want %q", decoded.Text, original.Text)
			}
		})
	}
}

// TestConcatManager 测试长短信管理器
func TestConcatManager(t *testing.T) {
	manager := NewConcatManager()

	// 添加 3 部分长短信
	for i := byte(1); i <= 3; i++ {
		msg := &Message{
			PhoneNumber: "+8613800138000",
			Text:        "Part " + string(rune('0'+i)),
			Reference:   0x42,
			Parts:       3,
			Part:        i,
		}

		result, err := manager.AddMessage(msg)
		if err != nil {
			t.Fatalf("AddMessage failed: %v", err)
		}

		if i < 3 {
			if result != nil {
				t.Error("Expected nil result for incomplete message")
			}
		} else {
			if result == nil {
				t.Fatal("Expected complete message")
			}
			if result.Text != "Part 1Part 2Part 3" {
				t.Errorf("Complete text mismatch: got %q", result.Text)
			}
		}
	}

	if manager.GetPendingCount() != 0 {
		t.Errorf("Expected 0 pending messages, got %d", manager.GetPendingCount())
	}
}

// TestConcatManagerSingleMessage 测试单条消息处理
func TestConcatManagerSingleMessage(t *testing.T) {
	manager := NewConcatManager()

	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        "Single message",
		Parts:       0,
	}

	result, err := manager.AddMessage(msg)
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected message to be returned")
	}

	if result.Text != msg.Text {
		t.Errorf("Text mismatch: got %q, want %q", result.Text, msg.Text)
	}
}

// TestUDHParseForConcatenatedSMS 测试长短信UDH解析
func TestUDHParseForConcatenatedSMS(t *testing.T) {
	// 这是一个真实的包含UDH的长短信PDU（第4部分，共4部分）
	pduStr := "0791448720006260400ED0E7B4D97C0E9BCD0000522140601451008E050003890404DC69FA5B0EAACFC3E7320BD40EBBC3E732680E2FBBC969F799059ADFD3F4311AE47ED3D3E6F4384C4FBFDD73D0DBFD7A9BCDA0B71C44AFCBDD20F93BDC4EBBCFA0B7D90C4ABB41F9775D0E0A8FC7EFBA9BAE039DD366B38B9D7F91C373B4F81D9693158A21BA5C96CF5DA069D85C06D1E5617B993D7701"

	msg, err := Decode(pduStr)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证UDH解析正确性
	if msg.Reference != 0x89 {
		t.Errorf("Reference错误: 期望 0x89, 得到 0x%02X", msg.Reference)
	}
	if msg.Parts != 4 {
		t.Errorf("Parts错误: 期望 4, 得到 %d", msg.Parts)
	}
	if msg.Part != 4 {
		t.Errorf("Part错误: 期望 4, 得到 %d", msg.Part)
	}

	// 确保文本解码成功
	if msg.Text == "" {
		t.Error("解码文本为空")
	}
}

// TestValidatePhoneNumber 测试电话号码验证
func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		number string
		valid  bool
	}{
		{"+8613800138000", true},
		{"13800138000", true},
		{"+1234", true},
		{"123", false},
		{"", false},
		{"+", false},
		{"abc123", false},
		{"123456789012345", true},
		{"1234567890123456", false},
	}

	for _, tt := range tests {
		t.Run(tt.number, func(t *testing.T) {
			result := ValidatePhoneNumber(tt.number)
			if result != tt.valid {
				t.Errorf("ValidatePhoneNumber(%q) = %v, want %v", tt.number, result, tt.valid)
			}
		})
	}
}

// TestIsGSM7BitCompatible 测试 GSM 7-bit 兼容性检查
func TestIsGSM7BitCompatible(t *testing.T) {
	tests := []struct {
		text       string
		compatible bool
	}{
		{"Hello World", true},
		{"Price: €10", true},
		{"|^€{}[]~\\", true},
		{"你好", false},
		{"Hello世界", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := IsGSM7BitCompatible(tt.text)
			if result != tt.compatible {
				t.Errorf("IsGSM7BitCompatible(%q) = %v, want %v", tt.text, result, tt.compatible)
			}
		})
	}
}

// TestCalculateMessageParts 测试消息分割计算
func TestCalculateMessageParts(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		encoding Encoding
		expected int
	}{
		{"Short 7-bit", "Hello", Encoding7Bit, 1},
		{"Long 7-bit", string(make([]byte, 200)), Encoding7Bit, 2},
		{"Short UCS2", "你好", EncodingUCS2, 1},
		{"Long UCS2", string(make([]rune, 100)), EncodingUCS2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMessageParts(tt.text, tt.encoding)
			if result != tt.expected {
				t.Errorf("CalculateMessageParts() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestGetMessageLength 测试消息长度计算
func TestGetMessageLength(t *testing.T) {
	tests := []struct {
		text     string
		encoding Encoding
		expected int
	}{
		{"Hello", Encoding7Bit, 5},
		{"€10", Encoding7Bit, 4}, // € 占 2 个字符
		{"你好", EncodingUCS2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := GetMessageLength(tt.text, tt.encoding)
			if result != tt.expected {
				t.Errorf("GetMessageLength(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

// TestMessageValidation 测试消息验证
func TestMessageValidation(t *testing.T) {
	tests := []struct {
		name    string
		msg     *Message
		wantErr bool
	}{
		{
			"Valid message",
			&Message{PhoneNumber: "+8613800138000", Text: "Hello", Type: MessageTypeSMSSubmit},
			false,
		},
		{
			"Empty phone",
			&Message{PhoneNumber: "", Text: "Hello", Type: MessageTypeSMSSubmit},
			true,
		},
		{
			"Empty text",
			&Message{PhoneNumber: "+8613800138000", Text: "", Type: MessageTypeSMSSubmit},
			true,
		},
		{
			"Invalid encoding",
			&Message{PhoneNumber: "+8613800138000", Text: "Hello", Encoding: 99},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFlashMessage 测试闪信
func TestFlashMessage(t *testing.T) {
	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        "Flash message",
		SMSC:        "+8613800138000",
		Flash:       true,
	}

	pdus, err := Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := Decode(pdus[0].Data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !decoded.Flash {
		t.Error("Flash flag not preserved")
	}
}

// TestValidityPeriod 测试有效期
func TestValidityPeriod(t *testing.T) {
	msg := &Message{
		PhoneNumber:    "+8613800138000",
		Text:           "Test",
		SMSC:           "+8613800138000",
		ValidityPeriod: ValidityPeriod24Hours,
	}

	pdus, err := Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if pdus[0].Data == "" {
		t.Error("PDU data is empty")
	}
}

// TestAutoReferenceGeneration 测试自动引用号生成
func TestAutoReferenceGeneration(t *testing.T) {
	longText := ""
	for i := 0; i < 200; i++ {
		longText += "a"
	}

	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        longText,
		SMSC:        "+8613800138000",
		Reference:   0, // 自动生成
	}

	pdus, err := Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(pdus) < 2 {
		t.Fatal("Expected multiple PDUs")
	}

	// 验证所有部分使用相同的引用号
	firstRef := byte(0)
	for i, pdu := range pdus {
		decoded, err := Decode(pdu.Data)
		if err != nil {
			t.Fatalf("Decode PDU %d failed: %v", i, err)
		}
		if i == 0 {
			firstRef = decoded.Reference
			if firstRef == 0 {
				t.Error("Reference was not generated")
			}
		} else {
			if decoded.Reference != firstRef {
				t.Errorf("Reference mismatch: PDU %d has %d, expected %d", i, decoded.Reference, firstRef)
			}
		}
	}
}

// TestSwapNibbles 测试半字节交换
func TestSwapNibbles(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1234", "2143"},
		{"12", "21"},
		{"123", "213"},
		{"", ""},
		{"1", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SwapNibbles(tt.input)
			if result != tt.expected {
				t.Errorf("SwapNibbles(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestHexConversion 测试十六进制转换
func TestHexConversion(t *testing.T) {
	data := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	hex := BytesToHex(data)
	expected := "0123456789ABCDEF"

	if hex != expected {
		t.Errorf("BytesToHex() = %q, want %q", hex, expected)
	}

	decoded, err := HexToBytes(hex)
	if err != nil {
		t.Fatalf("HexToBytes failed: %v", err)
	}

	for i, b := range decoded {
		if b != data[i] {
			t.Errorf("Byte %d mismatch: got %02X, want %02X", i, b, data[i])
		}
	}
}

// TestTimestampDecoding 测试时间戳解码
func TestTimestampDecoding(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		wantYear  int
		wantMonth time.Month
		wantDay   int
		wantHour  int
		wantMin   int
		wantSec   int
	}{
		{
			name:      "2020-08-26 19:37:41",
			timestamp: "02806291731480", // 来自真实 PDU
			wantYear:  2020,
			wantMonth: time.August,
			wantDay:   26,
			wantHour:  19,
			wantMin:   37,
			wantSec:   41,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeTimestamp(tt.timestamp)
			if err != nil {
				t.Fatalf("decodeTimestamp failed: %v", err)
			}

			if result.Year() != tt.wantYear {
				t.Errorf("Year mismatch: got %d, want %d", result.Year(), tt.wantYear)
			}
			if result.Month() != tt.wantMonth {
				t.Errorf("Month mismatch: got %v, want %v", result.Month(), tt.wantMonth)
			}
			if result.Day() != tt.wantDay {
				t.Errorf("Day mismatch: got %d, want %d", result.Day(), tt.wantDay)
			}
			if result.Hour() != tt.wantHour {
				t.Errorf("Hour mismatch: got %d, want %d", result.Hour(), tt.wantHour)
			}
			if result.Minute() != tt.wantMin {
				t.Errorf("Minute mismatch: got %d, want %d", result.Minute(), tt.wantMin)
			}
			if result.Second() != tt.wantSec {
				t.Errorf("Second mismatch: got %d, want %d", result.Second(), tt.wantSec)
			}
		})
	}
}
