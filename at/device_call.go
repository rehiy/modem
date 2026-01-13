package at

import "fmt"

// ===== 语音通话 =====

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

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CLIP: 1
		if label == "+CLIP" && len(param) >= 1 {
			status := parseInt(param[0])
			return status == 1, nil
		}
	}

	return false, fmt.Errorf("failed to parse caller ID status")
}

// SetCallerID 设置来电显示
func (m *Device) SetCallerID(enable bool) error {
	cmd := m.commands.CallerID
	if enable {
		cmd += "=1"
	} else {
		cmd += "=0"
	}
	return m.SendCommandExpect(cmd, "OK")
}

// CallInfo 通话信息
type CallInfo struct {
	ID     int    // 通话标识
	Dir    int    // 方向 [0: MO呼出, 1: MT呼入]
	Status int    // 状态 [0: 活动中, 1: 保持中, 2: 拨号中, 3: 响铃中, 4: 来电中]
	Mode   int    // 模式 [0: 语音, 1: 数据, 2: 传真]
	Number string // 号码
	Type   int    // 号码类型 [129: 国际, 161: 国内]
	Multip int    // 多方通话
}

// GetCallState 查询通话状态
func (m *Device) GetCallState() ([]CallInfo, error) {
	responses, err := m.SendCommand(m.commands.CallState)
	if err != nil {
		return nil, err
	}

	var calls []CallInfo
	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CLCC: 1,1,4,0,0,"10086",129
		if label == "+CLCC" && len(param) >= 7 {
			calls = append(calls, CallInfo{
				ID:     parseInt(param[0]),
				Dir:    parseInt(param[1]),
				Status: parseInt(param[2]),
				Mode:   parseInt(param[3]),
				Number: param[5],
				Type:   parseInt(param[6]),
				Multip: parseInt(param[4]),
			})
		}
	}

	if len(calls) == 0 {
		return nil, fmt.Errorf("no active calls")
	}
	return calls, nil
}

// GetCallWait 查询呼叫等待状态
func (m *Device) GetCallWait() (bool, error) {
	responses, err := m.SendCommand(m.commands.CallWait + "?")
	if err != nil {
		return false, err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CCWA: 0,1
		if label == "+CCWA" && len(param) >= 2 {
			status := parseInt(param[1])
			return status == 1, nil
		}
	}

	return false, fmt.Errorf("failed to parse call waiting status")
}

// SetCallWait 设置呼叫等待
func (m *Device) SetCallWait(enable bool) error {
	status := 0
	if enable {
		status = 1
	}
	cmd := fmt.Sprintf("%s=0,,%d", m.commands.CallWait, status)
	return m.SendCommandExpect(cmd, "OK")
}

// GetCallFWD 查询呼叫转移状态
// reason: 转移原因 [0: 无条件, 1: 遇忙, 2: 无应答, 3: 无法接通, 4: 所有]
func (m *Device) GetCallFWD(reason int) (bool, string, error) {
	responses, err := m.SendCommand(m.commands.CallFWD + fmt.Sprintf("=%d", reason))
	if err != nil {
		return false, "", err
	}

	for _, line := range responses {
		label, param := parseParam(line)
		// 格式: +CCFC: 0,3,"13800138000",145
		if label == "+CCFC" && len(param) >= 4 {
			status := parseInt(param[1])
			number := param[2]
			return status == 1, number, nil
		}
	}

	return false, "", fmt.Errorf("failed to parse call forward status")
}

// SetCallFWD 设置呼叫转移
// reason: 转移原因 [0: 无条件, 1: 遇忙, 2: 无应答, 3: 无法接通, 4: 所有]
// enable: 是否启用
// number: 转移号码
func (m *Device) SetCallFWD(reason int, enable bool, number string) error {
	status := 0
	if enable {
		status = 1
	}
	cmd := fmt.Sprintf("%s=%d,%d,\"%s\"", m.commands.CallFWD, reason, status, number)
	return m.SendCommandExpect(cmd, "OK")
}
