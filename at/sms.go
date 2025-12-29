package at

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf16"
)

// SMS 短信结构
type SMS struct {
	Index       int    // 短信索引
	Status      string // 短信状态：REC UNREAD, REC READ, STO UNSENT, STO SENT
	PhoneNumber string // 电话号码
	Timestamp   string // 时间戳
	Message     string // 短信内容
}

// LongSMS 长短信结构
type LongSMS struct {
	Reference uint8  // 长短信参考号
	Total     uint8  // 总段数
	Sequence  uint8  // 当前段序号
	Message   string // 当前段内容
}

// 短信最大长度
const (
	MaxSMSLength        = 160 // 英文短信最大长度
	MaxUCS2SMSLength    = 70  // UCS2编码短信最大长度
	MaxConcatSMSLength  = 153 // 英文长短信每段最大长度
	MaxUCS2ConcatLength = 67  // UCS2长短信每段最大长度
)

// SetSMSFormatText 设置短信格式为文本模式
func (m *Device) SetSMSFormatText() error {
	return m.SendCommandExpect(m.commands.SMSFormat+"=1", "OK")
}

// SetSMSFormatPDU 设置短信格式为 PDU 模式
func (m *Device) SetSMSFormatPDU() error {
	return m.SendCommandExpect(m.commands.SMSFormat+"=0", "OK")
}

// SendSMSText 发送文本短信（自动处理中文和长短信）
func (m *Device) SendSMSText(phoneNumber, message string) error {
	// 检查是否包含中文或其他非ASCII字符
	needsUCS2 := needsUCS2Encoding(message)

	// 判断是否需要分段发送
	maxLength := MaxSMSLength
	if needsUCS2 {
		maxLength = MaxUCS2SMSLength
	}

	// 如果消息长度超过限制，使用PDU模式发送长短信
	if len([]rune(message)) > maxLength {
		return m.sendLongSMS(phoneNumber, message, needsUCS2)
	}

	// 发送单条短信
	if needsUCS2 {
		// 使用PDU模式发送中文短信
		return m.sendUCS2SMS(phoneNumber, message)
	}

	// 发送普通文本短信
	return m.sendSimpleTextSMS(phoneNumber, message)
}

// SendSMSPDU 发送PDU格式短信
func (m *Device) SendSMSPDU(pduData string, length int) error {
	// 发送命令：AT+CMGS=length
	cmd := fmt.Sprintf("%s=%d", m.commands.SendSMS, length)
	fullCommand := cmd + "\r\n"

	return m.sendSMSCommand(fullCommand, pduData)
}

// ListSMS 列出所有短信
func (m *Device) ListSMS(status string) ([]SMS, error) {
	// status: "ALL", "REC UNREAD", "REC READ", "STO UNSENT", "STO SENT"
	cmd := fmt.Sprintf("%s=\"%s\"", m.commands.ListSMS, status)
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	return parseSMSList(responses), nil
}

// ReadSMS 读取指定索引的短信
func (m *Device) ReadSMS(index int) (*SMS, error) {
	cmd := fmt.Sprintf("%s=%d", m.commands.ReadSMS, index)
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	sms := parseSMS(responses)
	if sms == nil {
		return nil, fmt.Errorf("failed to parse SMS at index %d", index)
	}

	return sms, nil
}

// DeleteSMS 删除指定索引的短信
func (m *Device) DeleteSMS(index int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.DeleteSMS, index)
	return m.SendCommandExpect(cmd, "OK")
}

// DeleteAllSMS 删除所有短信
func (m *Device) DeleteAllSMS() error {
	// AT+CMGD=1,4 删除所有短信
	cmd := fmt.Sprintf("%s=1,4", m.commands.DeleteSMS)
	return m.SendCommandExpect(cmd, "OK")
}

// sendSMSCommand 通用的短信发送辅助函数
func (m *Device) sendSMSCommand(command string, data string) error {
	// 写入命令
	if err := m.writeString(command); err != nil {
		return fmt.Errorf("failed to write SMS command: %w", err)
	}

	// 发送数据，以 Ctrl+Z (0x1A) 结束
	dataWithCtrlZ := data + string(rune(0x1A))
	if err := m.writeString(dataWithCtrlZ); err != nil {
		return fmt.Errorf("failed to write SMS data: %w", err)
	}

	// 读取响应
	responses, err := m.readResponse()
	if err != nil {
		return fmt.Errorf("failed to read SMS response: %w", err)
	}

	// 检查是否成功
	hasSuccess := false
	for _, resp := range responses {
		if m.responses.IsSuccess(resp) {
			hasSuccess = true
			break
		}
	}
	if !hasSuccess {
		return fmt.Errorf("SMS send failed: %v", responses)
	}

	return nil
}

// sendSimpleTextSMS 发送简单文本短信（仅ASCII字符）
func (m *Device) sendSimpleTextSMS(phoneNumber, message string) error {
	// 发送命令：AT+CMGS="phoneNumber"
	cmd := fmt.Sprintf("%s=\"%s\"", m.commands.SendSMS, phoneNumber)
	fullCommand := cmd + "\r\n"

	// 等待 '>' 提示符
	// TODO: 实际应用中应该等待并检查 '>' 提示符

	return m.sendSMSCommand(fullCommand, message)
}

// sendUCS2SMS 发送UCS2编码的短信（支持中文）
func (m *Device) sendUCS2SMS(phoneNumber, message string) error {
	// 编码为UCS2
	ucs2Data := encodeUCS2(message)

	// 构建PDU数据
	pdu, length := buildPDU(phoneNumber, ucs2Data, 0, 0, 0)

	// 发送PDU短信
	return m.SendSMSPDU(pdu, length)
}

// sendLongSMS 发送长短信（自动分段）
func (m *Device) sendLongSMS(phoneNumber, message string, useUCS2 bool) error {
	// 生成长短信参考号（简单使用当前时间的低8位）
	// 实际应用中可以使用更复杂的算法
	reference := uint8(len(message) % 256)

	var segments []string
	var maxSegmentLength int

	if useUCS2 {
		maxSegmentLength = MaxUCS2ConcatLength
		// 将消息分段
		runes := []rune(message)
		for i := 0; i < len(runes); i += maxSegmentLength {
			end := i + maxSegmentLength
			if end > len(runes) {
				end = len(runes)
			}
			segments = append(segments, string(runes[i:end]))
		}
	} else {
		maxSegmentLength = MaxConcatSMSLength
		// 将消息分段
		for i := 0; i < len(message); i += maxSegmentLength {
			end := i + maxSegmentLength
			if end > len(message) {
				end = len(message)
			}
			segments = append(segments, message[i:end])
		}
	}

	totalSegments := uint8(len(segments))

	// 发送每一段
	for i, segment := range segments {
		sequence := uint8(i + 1)

		var pdu string
		var length int

		if useUCS2 {
			ucs2Data := encodeUCS2(segment)
			pdu, length = buildPDU(phoneNumber, ucs2Data, reference, totalSegments, sequence)
		} else {
			pdu, length = buildPDU(phoneNumber, segment, reference, totalSegments, sequence)
		}

		if err := m.SendSMSPDU(pdu, length); err != nil {
			return fmt.Errorf("failed to send segment %d/%d: %w", sequence, totalSegments, err)
		}
	}

	return nil
}

// needsUCS2Encoding 检查字符串是否需要UCS2编码
func needsUCS2Encoding(s string) bool {
	for _, r := range s {
		if r > 127 {
			return true
		}
	}
	return false
}

// encodeUCS2 将字符串编码为UCS2（UTF-16 BE）十六进制字符串
func encodeUCS2(s string) string {
	runes := []rune(s)
	utf16Codes := utf16.Encode(runes)

	var result strings.Builder
	for _, code := range utf16Codes {
		result.WriteString(fmt.Sprintf("%04X", code))
	}

	return result.String()
}

// decodeUCS2 将UCS2十六进制字符串解码为普通字符串
func decodeUCS2(hexStr string) (string, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}

	if len(data)%2 != 0 {
		return "", fmt.Errorf("invalid UCS2 data length")
	}

	utf16Codes := make([]uint16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		utf16Codes[i/2] = uint16(data[i])<<8 | uint16(data[i+1])
	}

	runes := utf16.Decode(utf16Codes)
	return string(runes), nil
}

// encodeBCD 将电话号码编码为BCD格式
func encodeBCD(phoneNumber string) string {
	// 如果号码长度为奇数，添加F
	if len(phoneNumber)%2 != 0 {
		phoneNumber += "F"
	}

	var result strings.Builder
	for i := 0; i < len(phoneNumber); i += 2 {
		// BCD编码：交换每对数字的位置
		result.WriteString(string(phoneNumber[i+1]))
		result.WriteString(string(phoneNumber[i]))
	}

	return result.String()
}

// buildPDU 构建PDU数据
// reference: 长短信参考号（0表示单条短信）
// total: 总段数（0表示单条短信）
// sequence: 当前段序号（0表示单条短信）
func buildPDU(phoneNumber, data string, reference, total, sequence uint8) (string, int) {
	var pdu strings.Builder

	// SMSC（使用默认，长度为0）
	pdu.WriteString("00")

	// PDU类型
	if total > 0 {
		// 长短信，包含用户数据头
		pdu.WriteString("51") // SMS-SUBMIT, UDHI=1
	} else {
		pdu.WriteString("11") // SMS-SUBMIT, UDHI=0
	}

	// 消息参考号（由设备自动分配）
	pdu.WriteString("00")

	// 目标号码长度和类型
	phoneLen := len(phoneNumber)
	if strings.HasPrefix(phoneNumber, "+") {
		phoneNumber = phoneNumber[1:]
		phoneLen = len(phoneNumber)
		pdu.WriteString(fmt.Sprintf("%02X", phoneLen))
		pdu.WriteString("91") // 国际格式
	} else {
		pdu.WriteString(fmt.Sprintf("%02X", phoneLen))
		pdu.WriteString("81") // 未知格式
	}

	// 编码电话号码（BCD格式）
	pdu.WriteString(encodeBCD(phoneNumber))

	// 协议标识
	pdu.WriteString("00")

	// 数据编码方案
	isUCS2 := len(data) > 0 && data[0] >= 'A' && data[0] <= 'F'
	if isUCS2 {
		pdu.WriteString("08") // UCS2编码
	} else {
		pdu.WriteString("00") // 7-bit编码
	}

	// 有效期（可选，这里省略）

	// 用户数据长度和内容
	if total > 0 {
		// 长短信，添加用户数据头
		udh := fmt.Sprintf("050003%02X%02X%02X", reference, total, sequence)
		udhLen := len(udh) / 2

		if isUCS2 {
			dataLen := len(data) / 2
			totalLen := udhLen + dataLen
			pdu.WriteString(fmt.Sprintf("%02X", totalLen))
			pdu.WriteString(udh)
			pdu.WriteString(data)
		} else {
			pdu.WriteString(fmt.Sprintf("%02X", len(data)+udhLen))
			pdu.WriteString(udh)
			pdu.WriteString(data)
		}
	} else {
		// 单条短信
		if isUCS2 {
			pdu.WriteString(fmt.Sprintf("%02X", len(data)/2))
			pdu.WriteString(data)
		} else {
			pdu.WriteString(fmt.Sprintf("%02X", len(data)))
			pdu.WriteString(data)
		}
	}

	// 计算TPDU长度（不包括SMSC部分）
	tpduLength := (len(pdu.String()) - 2) / 2

	return pdu.String(), tpduLength
}

// parseSMSList 解析短信列表
func parseSMSList(responses []string) []SMS {
	var smsList []SMS

	for i := 0; i < len(responses); i++ {
		line := responses[i]
		if strings.HasPrefix(line, "+CMGL:") {
			// 解析短信头
			sms := SMS{}
			parts := strings.Split(strings.TrimPrefix(line, "+CMGL:"), ",")
			if len(parts) >= 2 {
				fmt.Sscanf(parts[0], "%d", &sms.Index)
				sms.Status = trimQuotes(parts[1])
				if len(parts) >= 3 {
					sms.PhoneNumber = trimQuotes(parts[2])
				}
				if len(parts) >= 5 {
					sms.Timestamp = trimQuotes(parts[4])
				}
			}

			// 下一行是短信内容
			if i+1 < len(responses) {
				sms.Message = responses[i+1]
				i++ // 跳过内容行
			}

			smsList = append(smsList, sms)
		}
	}

	return smsList
}

// parseSMS 解析单条短信
func parseSMS(responses []string) *SMS {
	for i := 0; i < len(responses); i++ {
		line := responses[i]
		if strings.HasPrefix(line, "+CMGR:") {
			sms := &SMS{}
			parts := strings.Split(strings.TrimPrefix(line, "+CMGR:"), ",")
			if len(parts) >= 2 {
				sms.Status = trimQuotes(parts[0])
				sms.PhoneNumber = trimQuotes(parts[1])
				if len(parts) >= 4 {
					sms.Timestamp = trimQuotes(parts[3])
				}
			}

			// 下一行是短信内容
			if i+1 < len(responses) {
				sms.Message = responses[i+1]
			}

			return sms
		}
	}

	return nil
}

// trimQuotes 去除字符串两端的空格和引号
func trimQuotes(s string) string {
	return strings.Trim(strings.TrimSpace(s), "\"")
}
