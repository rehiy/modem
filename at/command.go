package at

import (
	"fmt"
	"strconv"
	"strings"
)

// CommandSet 定义可配置的 AT 命令集
type CommandSet struct {
	// 基本命令
	Test         string // 测试连接
	EchoOff      string // 关闭回显
	EchoOn       string // 开启回显
	Reset        string // 重置 modem
	FactoryReset string // 恢复出厂设置
	SaveSettings string // 保存设置

	// 信息查询
	Manufacturer string // 查询制造商
	Model        string // 查询型号
	Revision     string // 查询版本
	SerialNumber string // 查询序列号
	IMSI         string // 查询 IMSI
	ICCID        string // 查询 ICCID
	PhoneNumber  string // 查询手机号
	Operator     string // 查询运营商

	// 信号质量
	SignalQuality string // 查询信号质量

	// 网络注册
	NetworkRegistration string // 网络注册状态
	GPRSRegistration    string // GPRS 注册状态

	// 短信相关
	SMSFormat string // 设置短信格式
	ListSMS   string // 列出短信
	ReadSMS   string // 读取短信
	DeleteSMS string // 删除短信
	SendSMS   string // 发送短信

	// 通话相关
	Dial     string // 拨号
	Answer   string // 接听
	Hangup   string // 挂断
	CallerID string // 来电显示
}

// DefaultCommandSet 返回默认的标准 AT 命令集
func DefaultCommandSet() *CommandSet {
	return &CommandSet{
		// 基本命令
		Test:         "AT",
		EchoOff:      "ATE0",
		EchoOn:       "ATE1",
		Reset:        "ATZ",
		FactoryReset: "AT&F",
		SaveSettings: "AT&W",

		// 信息查询
		Manufacturer: "AT+CGMI",
		Model:        "AT+CGMM",
		Revision:     "AT+CGMR",
		SerialNumber: "AT+CGSN",
		IMSI:         "AT+CIMI",
		ICCID:        "AT+CCID",
		PhoneNumber:  "AT+CNUM",
		Operator:     "AT+COPS",

		// 信号质量
		SignalQuality: "AT+CSQ",

		// 网络注册
		NetworkRegistration: "AT+CREG",
		GPRSRegistration:    "AT+CGREG",

		// 短信相关
		SMSFormat: "AT+CMGF",
		ListSMS:   "AT+CMGL",
		ReadSMS:   "AT+CMGR",
		DeleteSMS: "AT+CMGD",
		SendSMS:   "AT+CMGS",

		// 通话相关
		Dial:     "ATD",
		Answer:   "ATA",
		Hangup:   "ATH",
		CallerID: "AT+CLIP",
	}
}

// ===== 基本命令 =====

// Test 测试连接
func (m *Device) Test() error {
	return m.SendCommandExpect(m.commands.Test, "OK")
}

// EchoOff 关闭回显
func (m *Device) EchoOff() error {
	return m.SendCommandExpect(m.commands.EchoOff, "OK")
}

// EchoOn 开启回显
func (m *Device) EchoOn() error {
	return m.SendCommandExpect(m.commands.EchoOn, "OK")
}

// Reset 重启模块
func (m *Device) Reset() error {
	return m.SendCommandExpect(m.commands.Reset, "OK")
}

// FactoryReset 恢复出厂设置
func (m *Device) FactoryReset() error {
	return m.SendCommandExpect(m.commands.FactoryReset, "OK")
}

// SaveSettings 保存设置
func (m *Device) SaveSettings() error {
	return m.SendCommandExpect(m.commands.SaveSettings, "OK")
}

// ===== 信息查询 =====

// SmpleQuery 通用简单信息查询函数
func (m *Device) SmpleQuery(command string) (string, error) {
	responses, err := m.SendCommand(command)
	if err != nil {
		return "", err
	}

	// 查找信息行（不以AT开头的行）
	for _, resp := range responses {
		if !strings.HasPrefix(resp, "AT") {
			return strings.TrimSpace(resp), nil
		}
	}

	return "", fmt.Errorf("no info found for command: %s", command)
}

// GetManufacturer 查询制造商信息
func (m *Device) GetManufacturer() (string, error) {
	return m.SmpleQuery(m.commands.Manufacturer)
}

// GetModel 查询型号信息
func (m *Device) GetModel() (string, error) {
	return m.SmpleQuery(m.commands.Model)
}

// GetRevision 查询版本信息
func (m *Device) GetRevision() (string, error) {
	return m.SmpleQuery(m.commands.Revision)
}

// GetSerialNumber 查询序列号
func (m *Device) GetSerialNumber() (string, error) {
	return m.SmpleQuery(m.commands.SerialNumber)
}

// GetIMSI 查询IMSI信息
func (m *Device) GetIMSI() (string, error) {
	return m.SmpleQuery(m.commands.IMSI)
}

// GetICCID 查询ICCID信息
func (m *Device) GetICCID() (string, error) {
	return m.SmpleQuery(m.commands.ICCID)
}

// GetPhoneNumber 查询手机号
func (m *Device) GetPhoneNumber() (string, error) {
	responses, err := m.SendCommand(m.commands.PhoneNumber)
	if err != nil {
		return "", err
	}

	for _, resp := range responses {
		if cnumData, ok := strings.CutPrefix(resp, "+CNUM:"); ok {
			// 格式: +CNUM: ,"+8613800138000",129
			parts := strings.Split(cnumData, ",")
			if len(parts) >= 2 {
				// 提取引号中的手机号
				number := strings.Trim(parts[1], `"'`)
				if number != "" {
					return number, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no phone number found")
}

// GetOperator 查询运营商信息
func (m *Device) GetOperator() (int, string, string, error) {
	responses, err := m.SendCommand(m.commands.Operator + "?")
	if err != nil {
		return 0, "", "", err
	}

	for _, resp := range responses {
		if copsData, ok := strings.CutPrefix(resp, "+COPS:"); ok {
			// 格式: +COPS: 0,0,"China Mobile",7
			parts := strings.Split(copsData, ",")
			if len(parts) >= 3 {
				mode := parseInt(parts[0])
				format := parseInt(strings.Trim(parts[1], `"'`))
				operator := strings.Trim(parts[2], `"'`)
				return mode, operator, fmt.Sprintf("%d", format), nil
			}
		}
	}

	return 0, "", "", fmt.Errorf("failed to parse operator info")
}

// ===== 信号质量 =====

// GetSignalQuality 查询信号质量
func (m *Device) GetSignalQuality() (int, int, error) {
	responses, err := m.SendCommand(m.commands.SignalQuality)
	if err != nil {
		return 0, 0, err
	}

	for _, resp := range responses {
		if csqData, ok := strings.CutPrefix(resp, "+CSQ:"); ok {
			parts := strings.Split(csqData, ",")
			if len(parts) >= 2 {
				rssi := parseInt(parts[0])
				ber := parseInt(parts[1])
				return rssi, ber, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("failed to parse signal quality")
}

// ===== 网络注册 =====

// GetNetworkStatus 查询网络注册状态
func (m *Device) GetNetworkStatus() (int, int, error) {
	responses, err := m.SendCommand(m.commands.NetworkRegistration + "?")
	if err != nil {
		return 0, 0, err
	}

	for _, resp := range responses {
		if cregData, ok := strings.CutPrefix(resp, "+CREG:"); ok {
			parts := strings.Split(cregData, ",")
			if len(parts) >= 2 {
				n := parseInt(parts[0])
				stat := parseInt(parts[1])
				return n, stat, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("failed to parse network status")
}

// GetGPRSStatus 查询GPRS注册状态
func (m *Device) GetGPRSStatus() (int, int, error) {
	responses, err := m.SendCommand(m.commands.GPRSRegistration + "?")
	if err != nil {
		return 0, 0, err
	}

	for _, resp := range responses {
		if cgregData, ok := strings.CutPrefix(resp, "+CGREG:"); ok {
			parts := strings.Split(cgregData, ",")
			if len(parts) >= 2 {
				n := parseInt(parts[0])
				stat := parseInt(parts[1])
				return n, stat, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("failed to parse GPRS status")
}

// ===== 通话相关 =====

// Dial 拨打电话
func (m *Device) Dial(number string) error {
	return m.SendCommandExpect(m.commands.Dial+number, "OK")
}

// Answer 接听电话
func (m *Device) Answer() error {
	return m.SendCommandExpect(m.commands.Answer, "OK")
}

// Hangup 挂断电话
func (m *Device) Hangup() error {
	return m.SendCommandExpect(m.commands.Hangup, "OK")
}

// GetCallerID 获取来电显示状态
func (m *Device) GetCallerID() (bool, error) {
	responses, err := m.SendCommand(m.commands.CallerID + "?")
	if err != nil {
		return false, err
	}

	for _, resp := range responses {
		if clipData, ok := strings.CutPrefix(resp, "+CLIP:"); ok {
			parts := strings.Split(clipData, ",")
			if len(parts) >= 1 {
				status := parseInt(parts[0])
				return status == 1, nil
			}
		}
	}

	return false, fmt.Errorf("failed to parse caller ID status")
}

// SetCallerID 设置来电显示
func (m *Device) SetCallerID(enable bool) error {
	command := m.commands.CallerID
	if enable {
		command += "=1"
	} else {
		command += "=0"
	}
	return m.SendCommandExpect(command, "OK")
}

// ===== 辅助工具 =====

// parseInt 解析整数
func parseInt(s string) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0 // 保持与原来相同的错误处理行为
	}
	return v
}
