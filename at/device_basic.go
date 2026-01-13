package at

import (
	"fmt"
)

// ===== 基本控制 =====

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

// SaveSettings 保存设置到当前配置文件
func (m *Device) SaveSettings() error {
	return m.SendCommandExpect(m.commands.SaveSettings, "OK")
}

// LoadProfile 加载指定配置文件
// profile [0: 默认配置, 1: 配置文件1, 2: 配置文件2]
func (m *Device) LoadProfile(profile int) error {
	cmd := fmt.Sprintf("%s%d", m.commands.LoadProfile, profile)
	return m.SendCommandExpect(cmd, "OK")
}

// SaveProfile 保存到指定配置文件
// profile [0: 默认配置, 1: 配置文件1, 2: 配置文件2]
func (m *Device) SaveProfile(profile int) error {
	cmd := fmt.Sprintf("%s%d", m.commands.SaveProfile, profile)
	return m.SendCommandExpect(cmd, "OK")
}

// ===== 设备状态 =====

// GetBatteryLevel 查询电池电量
func (m *Device) GetBatteryLevel() (int, int, error) {
	responses, err := m.SendCommand(m.commands.BatteryLevel)
	if err != nil {
		return 0, 0, err
	}

	param, err := parseResponse(m.commands.BatteryLevel, responses, 2)
	if err != nil {
		return 0, 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), nil
}

// GetDeviceTemp 查询设备温度
func (m *Device) GetDeviceTemp() (int, int, error) {
	responses, err := m.SendCommand(m.commands.DeviceTemp)
	if err != nil {
		return 0, 0, err
	}

	param, err := parseResponse(m.commands.DeviceTemp, responses, 2)
	if err != nil {
		return 0, 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), nil
}

// GetNetworkTime 查询网络时间
func (m *Device) GetNetworkTime() (string, error) {
	responses, err := m.SendCommand(m.commands.NetworkTime + "?")
	if err != nil {
		return "", err
	}

	param, err := parseResponse(m.commands.NetworkTime, responses, 1)
	if err != nil {
		return "", err
	}
	return param[0], nil
}

// SetTime 设置网络时间
// 格式: "YY/MM/DD,HH:MM:SS+TZ"
func (m *Device) SetTime(timeStr string) error {
	cmd := fmt.Sprintf("%s=\"%s\"", m.commands.SetTime, timeStr)
	return m.SendCommandExpect(cmd, "OK")
}

// ===== SIM 卡管理 =====

// GetSIMStatus 查询 SIM 卡状态
func (m *Device) GetSIMStatus() (string, error) {
	return m.SimpleQuery(m.commands.SIMStatus)
}

// VerifyPIN 验证 PIN 码
func (m *Device) VerifyPIN(pin string) error {
	cmd := fmt.Sprintf("%s=\"%s\"", m.commands.PINVerify, pin)
	return m.SendCommandExpect(cmd, "OK")
}

// ChangePIN 修改 PIN 码
func (m *Device) ChangePIN(oldPIN, newPIN string) error {
	cmd := fmt.Sprintf("%s=\"SC\",\"%s\",\"%s\"", m.commands.PINChange, oldPIN, newPIN)
	return m.SendCommandExpect(cmd, "OK")
}

// UnlockPIN 锁定/解锁 PIN
// enable [true: 启用, false: 禁用]
// pinType ["SC": SIM 卡 PIN]
func (m *Device) UnlockPIN(pinType string, enable bool, password string) error {
	status := 0
	if enable {
		status = 1
	}
	cmd := fmt.Sprintf("%s=\"%s\",%d,\"%s\"", m.commands.PINLock, pinType, status, password)
	return m.SendCommandExpect(cmd, "OK")
}

// ===== 设备身份信息 =====

// GetIMEI 查询 IMEI
func (m *Device) GetIMEI() (string, error) {
	return m.SimpleQuery(m.commands.IMEI)
}

// GetManufacturer 查询制造商信息
func (m *Device) GetManufacturer() (string, error) {
	return m.SimpleQuery(m.commands.Manufacturer)
}

// GetModel 查询型号信息
func (m *Device) GetModel() (string, error) {
	return m.SimpleQuery(m.commands.Model)
}

// GetRevision 查询版本信息
func (m *Device) GetRevision() (string, error) {
	return m.SimpleQuery(m.commands.Revision)
}

// GetIMSI 查询IMSI信息
func (m *Device) GetIMSI() (string, error) {
	return m.SimpleQuery(m.commands.IMSI)
}

// GetICCID 查询ICCID信息
func (m *Device) GetICCID() (string, error) {
	return m.SimpleQuery(m.commands.ICCID)
}

// GetNumber 查询手机号
func (m *Device) GetNumber() (string, int, error) {
	responses, err := m.SendCommand(m.commands.Number)
	if err != nil {
		return "", 0, err
	}

	param, err := parseResponse(m.commands.Number, responses, 2)
	if err != nil {
		return "", 0, err
	}
	return param[1], parseInt(param[2]), nil
}
