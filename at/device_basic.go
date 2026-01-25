package at

import (
	"fmt"
)

// ===== 基本控制 =====

// Test 测试连接
func (m *Device) Test() error {
	return m.SendExpect(m.commands.Test, "OK")
}

// EchoOff 关闭回显
func (m *Device) EchoOff() error {
	return m.SendExpect(m.commands.EchoOff, "OK")
}

// EchoOn 开启回显
func (m *Device) EchoOn() error {
	return m.SendExpect(m.commands.EchoOn, "OK")
}

// Reset 重启模块
func (m *Device) Reset() error {
	return m.SendExpect(m.commands.Reset, "OK")
}

// FactoryReset 恢复出厂设置
func (m *Device) FactoryReset() error {
	return m.SendExpect(m.commands.FactoryReset, "OK")
}

// SaveSettings 保存设置到当前配置文件
func (m *Device) SaveSettings() error {
	return m.SendExpect(m.commands.SaveSettings, "OK")
}

// LoadProfile 加载指定配置文件
// profile: 配置文件编号 [0: 默认配置, 1: 配置文件1, 2: 配置文件2]
func (m *Device) LoadProfile(profile int) error {
	cmd := fmt.Sprintf("%s%d", m.commands.LoadProfile, profile)
	return m.SendExpect(cmd, "OK")
}

// SaveProfile 保存到指定配置文件
// profile: 配置文件编号 [0: 默认配置, 1: 配置文件1, 2: 配置文件2]
func (m *Device) SaveProfile(profile int) error {
	cmd := fmt.Sprintf("%s%d", m.commands.SaveProfile, profile)
	return m.SendExpect(cmd, "OK")
}

// ===== 设备状态 =====

// GetBatteryLevel 查询电池电量及充电状态
func (m *Device) GetBatteryLevel() (int, int, error) {
	responses, err := m.SendCommand(m.commands.BatteryLevel)
	if err != nil {
		return 0, 0, err
	}

	// 响应格式: "+CBC: <bcs>,<bcl>"
	// bcs: 电池充电状态 [0: 未充电, 1: 充电中]
	// bcl: 电池电量级别 [0-100]
	param, err := parseResponse(m.commands.BatteryLevel, responses, 2)
	if err != nil {
		return 0, 0, err
	}
	return parseInt(param[0]), parseInt(param[1]), nil
}

// GetDeviceTemp 查询设备温度及状态
// 返回温度值和状态 [0: 正常, 1: 过热]
func (m *Device) GetDeviceTemp() (int, int, error) {
	responses, err := m.SendCommand(m.commands.DeviceTemp)
	if err != nil {
		return 0, 0, err
	}

	// 响应格式: "+CPMUTEMP: <temp>,<status>"
	// temp: 温度值
	// status: 状态 [0: 正常, 1: 过热]
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

	// 响应格式: "+CCLK: <time>"
	// time: 时间字符串，格式为 "YY/MM/DD,HH:MM:SS+TZ"
	param, err := parseResponse(m.commands.NetworkTime, responses, 1)
	if err != nil {
		return "", err
	}
	return param[0], nil
}

// SetTime 设置网络时间
// timeStr: 时间字符串，格式为 "YY/MM/DD,HH:MM:SS+TZ"，例如 "26/01/15,14:30:00+08"
func (m *Device) SetTime(timeStr string) error {
	cmd := fmt.Sprintf("%s=\"%s\"", m.commands.SetTime, timeStr)
	return m.SendExpect(cmd, "OK")
}

// ===== SIM 卡管理 =====

// GetSIMStatus 查询 SIM 卡状态
// 返回 SIM 状态代码 ["READY": 准备就绪, "SIM PIN": 需要 PIN 码, "SIM PUK": 需要 PUK 码, "PH-SIM PIN": 需要 PH-SIM PIN]
func (m *Device) GetSIMStatus() (string, error) {
	// 响应格式: "+CPIN: <code>"
	// code: 状态代码 ["READY", "SIM PIN", "SIM PUK", "PH-SIM PIN"]
	return m.SimpleQuery(m.commands.SIMStatus)
}

// VerifyPIN 验证 PIN 码
// pin: PIN 码
func (m *Device) VerifyPIN(pin string) error {
	cmd := fmt.Sprintf("%s=\"%s\"", m.commands.PINVerify, pin)
	return m.SendExpect(cmd, "OK")
}

// ChangePIN 修改 PIN 码
// oldPIN: 旧 PIN 码
// newPIN: 新 PIN 码
func (m *Device) ChangePIN(oldPIN, newPIN string) error {
	cmd := fmt.Sprintf("%s=\"SC\",\"%s\",\"%s\"", m.commands.PINChange, oldPIN, newPIN)
	return m.SendExpect(cmd, "OK")
}

// UnlockPIN 设置 PIN 锁状态
// pinType: PIN 锁类型 ["SC": SIM 卡 PIN, "PS": SIM 卡 PUK, "PF": SIM 卡 FDN, "SC": 电话簿]
// enable: 是否启用 PIN 锁 [true: 启用, false: 禁用]
// password: PIN 码或 PUK 码
func (m *Device) UnlockPIN(pinType string, enable bool, password string) error {
	status := 0
	if enable {
		status = 1
	}
	cmd := fmt.Sprintf("%s=\"%s\",%d,\"%s\"", m.commands.PINLock, pinType, status, password)
	return m.SendExpect(cmd, "OK")
}

// ===== 设备身份信息 =====

// GetIMEI 查询 IMEI
func (m *Device) GetIMEI() (string, error) {
	// 响应格式: "<imei>"
	// imei: 15位设备唯一标识码
	return m.SimpleQuery(m.commands.IMEI)
}

// GetManufacturer 查询制造商信息
func (m *Device) GetManufacturer() (string, error) {
	// 响应格式: "<manufacturer>"
	// manufacturer: 制造商名称
	return m.SimpleQuery(m.commands.Manufacturer)
}

// GetModel 查询型号信息
func (m *Device) GetModel() (string, error) {
	// 响应格式: "<model>"
	// model: 设备型号
	return m.SimpleQuery(m.commands.Model)
}

// GetRevision 查询版本信息
func (m *Device) GetRevision() (string, error) {
	// 响应格式: "<revision>"
	// revision: 固件版本号
	return m.SimpleQuery(m.commands.Revision)
}

// GetIMSI 查询IMSI信息
func (m *Device) GetIMSI() (string, error) {
	// 响应格式: "<imsi>"
	// imsi: 15位国际移动用户识别码
	return m.SimpleQuery(m.commands.IMSI)
}

// GetICCID 查询ICCID信息
func (m *Device) GetICCID() (string, error) {
	// 响应格式: "<iccid>"
	// iccid: 20位集成电路卡识别码
	return m.SimpleQuery(m.commands.ICCID)
}

// GetNumber 查询本机号码
// 返回 (电话号码, 号码类型)
func (m *Device) GetNumber() (string, int, error) {
	responses, err := m.SendCommand(m.commands.Number)
	if err != nil {
		return "", 0, err
	}

	// 响应格式: "+CNUM: <alpha>,<number>,<type>"
	// alpha: 名称
	// number: 电话号码
	// type: 号码类型 [129: 国际, 161: 国内]
	param, err := parseResponse(m.commands.Number, responses, 2)
	if err != nil {
		return "", 0, err
	}
	return param[1], parseInt(param[2]), nil
}
