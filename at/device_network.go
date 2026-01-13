package at

import "fmt"

// ===== 网络状态 =====

// GetOperator 查询运营商信息
func (m *Device) GetOperator() (int, int, string, int, error) {
	responses, err := m.SendCommand(m.commands.Operator + "?")
	if err != nil {
		return 0, 0, "", 0, err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +COPS: 0,2,"46001",7
		if label == "+COPS" && len(param) >= 3 {
			mode := parseInt(param[0])
			format := parseInt(param[1])
			oper := param[2]          // 运营商
			act := parseInt(param[3]) // 接入技术
			return mode, format, oper, act, nil
		}
	}

	return 0, 0, "", 0, fmt.Errorf("failed to parse operator info")
}

// GetNetworkMode 查询网络模式
func (m *Device) GetNetworkMode() (int, error) {
	responses, err := m.SendCommand(m.commands.NetworkMode + "?")
	if err != nil {
		return 0, err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CNMP: 38
		if label == "+CNMP" && len(param) >= 1 {
			return parseInt(param[0]), nil
		}
	}

	return 0, fmt.Errorf("failed to parse network mode")
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

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CREG: 0,1
		if label == "+CREG" && len(param) >= 2 {
			return parseInt(param[0]), parseInt(param[1]), nil
		}
	}

	return 0, 0, fmt.Errorf("failed to parse network status")
}

// GetGPRSStatus 查询GPRS注册状态
func (m *Device) GetGPRSStatus() (int, int, error) {
	responses, err := m.SendCommand(m.commands.GPRSReg + "?")
	if err != nil {
		return 0, 0, err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CGREG: 0,1
		if label == "+CGREG" && len(param) >= 2 {
			return parseInt(param[0]), parseInt(param[1]), nil
		}
	}

	return 0, 0, fmt.Errorf("failed to parse GPRS status")
}

// GetSignalQuality 查询信号质量
func (m *Device) GetSignalQuality() (int, int, error) {
	responses, err := m.SendCommand(m.commands.Signal)
	if err != nil {
		return 0, 0, err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CSQ: 15,0
		if label == "+CSQ" && len(param) >= 2 {
			rssi := parseInt(param[0])
			ber := parseInt(param[1])
			return rssi, ber, nil
		}
	}

	return 0, 0, fmt.Errorf("failed to parse signal quality")
}

// ===== 网络配置 =====

// GetAPN 查询 APN 配置
func (m *Device) GetAPN(cid int) (int, string, string, error) {
	responses, err := m.SendCommand(m.commands.APN + "?")
	if err != nil {
		return 0, "", "", err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CGDCONT: 1,"IP","cmnet","","0.0.0.0",0,0
		if label == "+CGDCONT" && len(param) >= 3 {
			configCID := parseInt(param[0])
			if cid != 0 && configCID != cid {
				return 0, "", "", fmt.Errorf("APN with cid %d not found", cid)
			}
			return configCID, param[1], param[2], nil
		}
	}

	return 0, "", "", fmt.Errorf("failed to parse APN configuration")
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

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CGACT: 1,1
		if label == "+CGACT" && len(param) >= 2 {
			contextCID := parseInt(param[0])
			if cid != 0 && contextCID != cid {
				continue
			}
			return contextCID, parseInt(param[1]), nil
		}
	}

	return 0, 0, fmt.Errorf("failed to parse PDP context status")
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

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CGPADDR: 1,"10.1.2.3"
		if label == "+CGPADDR" && len(param) >= 2 {
			configCID := parseInt(param[0])
			if cid != 0 && configCID != cid {
				return 0, "", fmt.Errorf("IP address with cid %d not found", cid)
			}
			return configCID, param[1], nil
		}
	}

	return 0, "", fmt.Errorf("failed to parse IP address")
}

// ===== 通知管理 =====

// GetNetworkRegNotify 查询网络注册通知状态
func (m *Device) GetNetworkRegNotify() (int, error) {
	responses, err := m.SendCommand(m.commands.NetworkRegNotify + "?")
	if err != nil {
		return 0, err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CREG: 2,1
		if label == "+CREG" && len(param) >= 1 {
			return parseInt(param[0]), nil
		}
	}

	return 0, fmt.Errorf("failed to parse network reg notify status")
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

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CGREG: 2,1
		if label == "+CGREG" && len(param) >= 1 {
			return parseInt(param[0]), nil
		}
	}

	return 0, fmt.Errorf("failed to parse GPRS reg notify status")
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
