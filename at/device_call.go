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

	// 响应格式: "+CLIP: <n>"
	// n: 来电显示状态 [0: 禁用, 1: 启用]
	param, err := parseResponse(m.commands.CallerID, responses, 1)
	if err != nil {
		return false, err
	}
	return parseInt(param[0]) == 1, nil
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
	label := getCommandResponseLabel(m.commands.CallState)
	for _, line := range responses {
		respLabel, param := parseParam(line)
		if respLabel == label && len(param) >= 7 {
			// 响应格式: "+CLCC: <id>,<dir>,<status>,<mode>,<multip>,<number>,<type>"
			// id: 通话标识
			// dir: 方向 [0: MO呼出, 1: MT呼入]
			// status: 状态 [0: 活动中, 1: 保持中, 2: 拨号中, 3: 响铃中, 4: 来电中]
			// mode: 模式 [0: 语音, 1: 数据, 2: 传真]
			// multip: 多方通话
			// number: 号码
			// type: 号码类型 [129: 国际, 161: 国内]
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

	// 响应格式: "+CCWA: <status>,<class1>,[<class2>,...]"
	// status: 呼叫等待状态 [0: 禁用, 1: 启用]
	// class: 通话类型 [1: 语音, 2: 数据, 4: 传真, 7: 所有]
	param, err := parseResponse(m.commands.CallWait, responses, 2)
	if err != nil {
		return false, err
	}
	return parseInt(param[1]) == 1, nil
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

	// 响应格式: "+CCFC: <status>,<class>,<number>,<type>"
	// status: 状态 [0: 禁用, 1: 启用]
	// class: 通话类型 [1: 语音, 2: 数据, 4: 传真, 7: 所有]
	// number: 转移号码
	// type: 号码类型 [129: 国际, 161: 国内]
	param, err := parseResponse(m.commands.CallFWD, responses, 4)
	if err != nil {
		return false, "", err
	}
	return parseInt(param[1]) == 1, param[2], nil
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
