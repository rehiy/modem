package at

import "fmt"

// ===== 网络状态 =====

// GetOperator 查询运营商信息
func (m *Device) GetOperator() (int, int, string, int, error) {
	responses, err := m.SendCommand(m.commands.Operator + "?")
	if err != nil {
		return 0, 0, "", 0, err
	}

	param, err := parseResponse(m.commands.Operator, responses, 3)
	if err != nil {
		return 0, 0, "", 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), param[2], parseInt(param[3]), nil
}

// GetNetworkMode 查询网络模式
func (m *Device) GetNetworkMode() (int, error) {
	responses, err := m.SendCommand(m.commands.NetworkMode + "?")
	if err != nil {
		return 0, err
	}

	param, err := parseResponse(m.commands.NetworkMode, responses, 1)
	if err != nil {
		return 0, err
	}
	return parseInt(param[0]), nil
}

// SetNetworkMode 设置网络模式
// 常用模式: 2=AUTOMATIC, 13=GSM ONLY, 38=LTE ONLY, 51=SA/NSA
func (m *Device) SetNetworkMode(mode int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.NetworkMode, mode)
	return m.SendCommandExpect(cmd, "OK")
}

// GetNetworkStatus 查询网络注册状态
func (m *Device) GetNetworkStatus() (int, int, error) {
	responses, err := m.SendCommand(m.commands.NetworkReg + "?")
	if err != nil {
		return 0, 0, err
	}

	param, err := parseResponse(m.commands.NetworkReg, responses, 2)
	if err != nil {
		return 0, 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), nil
}

// GetGPRSStatus 查询GPRS注册状态
func (m *Device) GetGPRSStatus() (int, int, error) {
	responses, err := m.SendCommand(m.commands.GPRSReg + "?")
	if err != nil {
		return 0, 0, err
	}

	param, err := parseResponse(m.commands.GPRSReg, responses, 2)
	if err != nil {
		return 0, 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), nil
}

// GetSignalQuality 查询信号质量
func (m *Device) GetSignalQuality() (int, int, error) {
	responses, err := m.SendCommand(m.commands.Signal)
	if err != nil {
		return 0, 0, err
	}

	param, err := parseResponse(m.commands.Signal, responses, 2)
	if err != nil {
		return 0, 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), nil
}

// ===== 网络配置 =====

// GetAPN 查询 APN 配置
func (m *Device) GetAPN(cid int) (int, string, string, error) {
	responses, err := m.SendCommand(m.commands.APN + "?")
	if err != nil {
		return 0, "", "", err
	}

	filter := func(param map[int]string) bool {
		return cid == 0 || parseInt(param[0]) == cid
	}

	param, err := parseResponseFiltered(m.commands.APN, responses, 3, filter)
	if err != nil {
		return 0, "", "", err
	}
	return parseInt(param[0]), param[1], param[2], nil
}

// SetAPN 设置 APN 配置
// cid: 上下文标识符 [1-]
// pdpType: PDP 类型 ["IP", "IPV6", "IPV4V6"]
// apn: 接入点名称
func (m *Device) SetAPN(cid int, pdpType, apn string) error {
	cmd := fmt.Sprintf("%s=%d,\"%s\",\"%s\"", m.commands.SetAPN, cid, pdpType, apn)
	return m.SendCommandExpect(cmd, "OK")
}

// GetPDPContext 查询 PDP 上下文状态
// cid: 上下文标识符
func (m *Device) GetPDPContext(cid int) (int, int, error) {
	responses, err := m.SendCommand(m.commands.PDPContext + "?")
	if err != nil {
		return 0, 0, err
	}

	filter := func(param map[int]string) bool {
		return cid == 0 || parseInt(param[0]) == cid
	}

	param, err := parseResponseFiltered(m.commands.PDPContext, responses, 2, filter)
	if err != nil {
		return 0, 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), nil
}

// SetPDPContext 设置 PDP 上下文状态
// cid: 上下文标识符
// state: 状态 [0: 停用, 1: 激活]
func (m *Device) SetPDPContext(cid int, state int) error {
	cmd := fmt.Sprintf("%s=%d,%d", m.commands.PDPContext, cid, state)
	return m.SendCommandExpect(cmd, "OK")
}

// GetIPAddress 查询 IP 地址
func (m *Device) GetIPAddress(cid int) (int, string, error) {
	responses, err := m.SendCommand(m.commands.IPAddress + "?")
	if err != nil {
		return 0, "", err
	}

	filter := func(param map[int]string) bool {
		return cid == 0 || parseInt(param[0]) == cid
	}

	param, err := parseResponseFiltered(m.commands.IPAddress, responses, 2, filter)
	if err != nil {
		return 0, "", err
	}
	return parseInt(param[0]), param[1], nil
}

// ===== 通知管理 =====

// GetNetworkRegNotify 查询网络注册通知状态
func (m *Device) GetNetworkRegNotify() (int, error) {
	responses, err := m.SendCommand(m.commands.NetworkRegNotify + "?")
	if err != nil {
		return 0, err
	}

	param, err := parseResponse(m.commands.NetworkRegNotify, responses, 1)
	if err != nil {
		return 0, err
	}
	return parseInt(param[0]), nil
}

// SetNetworkRegNotify 设置网络注册通知
// mode [0: 禁用, 1: 启用, 2: 启用并显示位置信息]
func (m *Device) SetNetworkRegNotify(mode int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.NetworkRegNotify, mode)
	return m.SendCommandExpect(cmd, "OK")
}

// GetGPRSRegNotify 查询 GPRS 注册通知状态
func (m *Device) GetGPRSRegNotify() (int, error) {
	responses, err := m.SendCommand(m.commands.GPRSRegNotify + "?")
	if err != nil {
		return 0, err
	}

	param, err := parseResponse(m.commands.GPRSRegNotify, responses, 1)
	if err != nil {
		return 0, err
	}
	return parseInt(param[0]), nil
}

// SetGPRSRegNotify 设置 GPRS 注册通知
// mode [0: 禁用, 1: 启用, 2: 启用并显示位置信息]
func (m *Device) SetGPRSRegNotify(mode int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.GPRSRegNotify, mode)
	return m.SendCommandExpect(cmd, "OK")
}

// SetSignalReport 设置信号质量上报
// mode [0: 关闭, 1: 开启]
// interval: 上报间隔(秒) [1-255]
func (m *Device) SetSignalReport(mode int, interval int) error {
	cmd := fmt.Sprintf("%s=%d,%d", m.commands.SignalReport, mode, interval)
	return m.SendCommandExpect(cmd, "OK")
}
