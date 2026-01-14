package at

import (
	"fmt"
	"sort"
	"time"

	"github.com/rehiy/modem/sms"
	"github.com/rehiy/modem/sms/pdumode"
)

// SMS 短信信息
type Sms struct {
	Number  string `json:"number"`
	Text    string `json:"text"`
	Time    string `json:"time"`
	Index   int    `json:"index"`   // 首个分片的索引
	Indices []int  `json:"indices"` // 所有分片的索引
	Status  string `json:"status"`  // 短信状态 [PDU: TEXT, 0: "REC UNREAD", 1: "REC READ", 2: "STO UNSENT", 3: "STO SENT", 4: "ALL"]
}

// SetSmsMode 设置短信模式
// v [0: PDU 模式, 1: TEXT 模式]
func (m *Device) SetSmsMode(v int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.SmsFormat, v)
	return m.SendCommandExpect(cmd, "OK")
}

// GetSmsMode 获取短信模式
// 返回 [0: PDU 模式, 1: TEXT 模式]
func (m *Device) GetSmsMode() (int, error) {
	responses, err := m.SendCommand(m.commands.SmsFormat + "?")
	if err != nil {
		return 0, err
	}

	// 响应格式: "+CMGF: <mode>"
	param, err := parseResponse(m.commands.SmsFormat+"?", responses, 1)
	if err != nil {
		return 0, err
	}

	return parseInt(param[0]), nil
}

// SetSmsStore 设置短信存储
// v [ME: 手机内存, SM: 短信存储]
func (m *Device) SetSmsStore(v1, v2, v3 string) error {
	cmd := fmt.Sprintf("%s=\"%s\",\"%s\",\"%s\"", m.commands.SmsStore, v1, v2, v3)
	return m.SendCommandExpect(cmd, "OK")
}

// GetSmsStore 获取短信存储配置
// 返回 (读存储, 写存储, 接收存储)
func (m *Device) GetSmsStore() (map[string]any, error) {
	responses, err := m.SendCommand(m.commands.SmsStore + "?")
	if err != nil {
		return nil, err
	}

	// 响应格式: "+CPMS: <mem1>,<used1>,<total1>,<mem2>,<used2>,<total2>,<mem3>,<used3>,<total3>"
	// mem1/2/3: 读取/写入/接收短信的存储位置
	param, err := parseResponse(m.commands.SmsStore+"?", responses, 9)
	if err != nil {
		return nil, err
	}

	result := map[string]any{
		"mem1":   param[0],
		"used1":  parseInt(param[1]),
		"total1": parseInt(param[2]),
		"mem2":   param[3],
		"used2":  parseInt(param[4]),
		"total2": parseInt(param[5]),
		"mem3":   param[6],
		"used3":  parseInt(param[7]),
		"total3": parseInt(param[8]),
	}
	return result, nil
}

// GetSmsCenter 获取短信中心号码
func (m *Device) GetSmsCenter() (string, int, error) {
	responses, err := m.SendCommand(m.commands.SmsCenter + "?")
	if err != nil {
		return "", 0, err
	}

	// 响应格式: "+CSCA: <number>,<tosca>"
	// number: 短信中心号码
	// tosca: 号码类型
	param, err := parseResponse(m.commands.SmsCenter+"?", responses, 2)
	if err != nil {
		return "", 0, err
	}

	m.printf("param: %v", param)

	return param[0], parseInt(param[1]), nil
}

// SetSmsCenter 设置短信中心号码
func (m *Device) SetSmsCenter(number string) error {
	cmd := fmt.Sprintf("%s=\"%s\"", m.commands.SmsCenter, number)
	return m.SendCommandExpect(cmd, "OK")
}

// SendSmsPdu 发送短信
func (m *Device) SendSmsPdu(number, message string) error {
	tpdus, err := sms.Encode([]byte(message), sms.To(number))
	if err != nil {
		return err
	}

	// 临时延长超时
	rdTimeout := m.timeout
	m.timeout = time.Second * 60
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
		cmd := fmt.Sprintf("%s=%d", m.commands.SendSms, len(tpduBytes))
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

// ListSMSPdu 获取短信列表
func (m *Device) ListSmsPdu(stat int) ([]Sms, error) {
	cmd := fmt.Sprintf("%s=%d", m.commands.ListSms, stat)
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	result := []Sms{}
	indices := make(map[int][]int)
	collector := sms.NewCollector()
	defer collector.Close() // 确保资源释放

	// 响应格式: "+CMGL: <index>,<stat>,[<alpha>],<length>"
	// index: 短信索引
	// stat: 状态 [0: REC UNREAD, 1: REC READ, 2: STO UNSENT, 3: STO SENT]
	// alpha: 发送者名称
	// length: 长度
	// 下一行: PDU 十六进制数据
	expectedLabel := getCommandResponseLabel(m.commands.ListSms)
	for i, l := 0, len(responses); i < l; {
		label, param := parseParam(responses[i])
		i++

		if label != expectedLabel || len(param) < 2 {
			continue
		}

		// 无下一行，退出
		if i >= l {
			break
		}

		// 提取 PDU 数据
		pduHex := responses[i]
		i++

		// 解析十六进制 PDU
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
		if mref == 0 {
			mref = index
		}
		indices[mref] = append(indices[mref], index)

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

			result = append(result, Sms{
				Number:  segments[0].OA.Number(),
				Text:    string(msgBytes),
				Time:    segments[0].SCTS.Time.Format("2006/01/02 15:04:05"),
				Index:   indices[mref][0],
				Indices: indices[mref],
				Status:  param[1],
			})
			delete(indices, mref)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Index > result[j].Index
	})
	return result, nil
}

// DeleteSms 批量删除指定索引的短信
func (m *Device) DeleteSms(indices []int) error {
	for _, index := range indices {
		cmd := fmt.Sprintf("%s=%d", m.commands.DeleteSms, index)
		if _, err := m.SendCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}
