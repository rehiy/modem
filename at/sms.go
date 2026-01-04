package at

import (
	"fmt"
	"sort"
	"time"

	"github.com/rehiy/modem/sms"
	"github.com/rehiy/modem/sms/pdumode"
)

// SMS 短信信息
type SMS struct {
	PhoneNumber string `json:"phoneNumber"`
	Text        string `json:"text"`
	Time        string `json:"time"`
	Index       int    `json:"index"`   // 首个分片的索引
	Indices     []int  `json:"indices"` // 所有分片的索引
	Status      string `json:"status"`  // 短信状态 [PDU: TEXT, 0: "REC UNREAD", 1: "REC READ", 2: "STO UNSENT", 3: "STO SENT", 4: "ALL"]
}

// SetSMSMode 设置短信模式
// v [0: PDU 模式, 1: TEXT 模式]
func (m *Device) SetSMSMode(v int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.SMSFormat, v)
	return m.SendCommandExpect(cmd, "OK")
}

// ListSMSPdu 获取短信列表
func (m *Device) ListSMSPdu(stat int) ([]SMS, error) {
	cmd := fmt.Sprintf("%s=%d", m.commands.ListSMS, stat)
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	result := []SMS{}
	collector := sms.NewCollector()
	indices := make(map[int][]int) // refKey -> 所有分片索引

	for i, l := 0, len(responses); i < l; {
		label, param := parseParam(responses[i])
		i++

		if label != "+CMGL" || len(param) < 2 {
			continue
		}

		// 无下一行，退出
		if i >= l {
			break
		}

		// 解码 PDU 数据
		pduHex := responses[i]
		i++

		// 使用 pdumode 解析十六进制 PDU 字符串
		pdu, err := pdumode.UnmarshalHexString(pduHex)
		if err != nil {
			m.printf("unmarshal pdu error: %v", err)
			continue
		}

		// 从 PDU 中解析 TPDU
		tpduMsg, err := sms.Unmarshal(pdu.TPDU)
		if err != nil {
			m.printf("unmarshal tpdu error: %v", err)
			continue
		}

		// 记录索引和引用号
		index := parseInt(param[0])
		_, _, mref, _ := tpduMsg.ConcatInfo()
		refKey := int(mref)
		// 长短信用 mref 作为 key，短短信用 index 作为 key
		if refKey == 0 {
			refKey = index
		}
		indices[refKey] = append(indices[refKey], index)

		// 收集短信（长短信自动合并）
		segments, err := collector.Collect(*tpduMsg)
		if err != nil {
			m.printf("collect sms %d error: %v", index, err)
			continue
		}

		// 收集到完整短信时解码并添加
		if len(segments) > 0 {
			msgBytes, err := sms.Decode(segments)
			if err != nil {
				m.printf("decode sms error: %v", err)
				continue
			}

			result = append(result, SMS{
				PhoneNumber: segments[0].OA.Number(),
				Text:        string(msgBytes),
				Time:        segments[0].SCTS.Time.Format("2006/01/02 15:04:05"),
				Index:       indices[refKey][0],
				Indices:     indices[refKey],
				Status:      param[1],
			})
			delete(indices, refKey)
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Index < result[j].Index })
	return result, nil
}

// SendSMSPdu 发送短信
func (m *Device) SendSMSPdu(number, message string) error {
	tpdus, err := sms.Encode([]byte(message), sms.To(number))
	if err != nil {
		return err
	}

	// 临时延长超时
	rdTimeout := m.timeout
	m.timeout = time.Second * 15
	defer func() { m.timeout = rdTimeout }()

	for _, p := range tpdus {
		// 将 TPDU 序列化为字节数组
		tpduBytes, err := p.MarshalBinary()
		if err != nil {
			m.printf("marshal tpdu error: %v", err)
			return err
		}

		// 使用 pdumode 包装 TPDU 并编码为十六进制
		pdu := &pdumode.PDU{TPDU: tpduBytes}
		pduHex, err := pdu.MarshalHexString()
		if err != nil {
			m.printf("marshal pdu error: %v", err)
			return err
		}

		// 发送 AT 命令（TPDU 长度不包含 SMSC 部分）
		cmd := fmt.Sprintf("%s=%d", m.commands.SendSMS, len(tpduBytes))
		if err := m.SendCommandExpect(cmd, ">"); err != nil {
			m.printf("send sms command error: %v", err)
			return err
		}

		// 发送 PDU 数据
		if _, err := m.SendCommand(pduHex + "\x1A"); err != nil {
			m.printf("send sms response error: %v", err)
			return err
		}
	}

	return nil
}

// DeleteSMS 删除指定索引的短信
func (m *Device) DeleteSMS(index int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.DeleteSMS, index)
	_, err := m.SendCommand(cmd)
	return err
}
