package at

import (
	"strings"
)

// NotificationSet 定义可配置的URC（Unsolicited Result Code）类型集合
type NotificationSet struct {
	// 通话相关
	Ring              string // 来电响铃
	CallRing          string // 来电铃声类型
	CallerID          string // 来电显示
	CallWaiting       string // 呼叫等待
	ConnectedLine     string // 连接线号呈现
	SuppressNotify    string // 补充服务通知（失败）
	UnsolicitedNotify string // 补充服务通知（成功）

	// 短信相关
	SMSReady        string // 新短信到达通知
	SMSContent      string // 短信内容推送
	SMSStatusReport string // 短信状态报告
	CellBroadcast   string // 小区广播消息

	// 网络注册相关
	NetworkReg string // GSM 网络注册状态
	GPRSReg    string // GPRS 网络注册状态
	EPSReg     string // EPS（4G）网络注册状态

	// 服务相关
	USSD string // 非结构化补充业务数据

	// 状态变化
	StatusChange string // 设备状态变化

	// 连接相关
	DataCallStatus string // 数据呼叫状态

	// 时区和网络
	NetworkTime string // 网络时间更新（NITZ）
}

// DefaultNotificationSet 返回默认的URC类型集合
func DefaultNotificationSet() *NotificationSet {
	return &NotificationSet{
		// 通话相关
		Ring:              "RING",
		CallRing:          "+CRING:",
		CallerID:          "+CLIP:",
		CallWaiting:       "+CCWA:",
		ConnectedLine:     "+COLP:",
		SuppressNotify:    "+CSSI:",
		UnsolicitedNotify: "+CSSU:",

		// 短信相关
		SMSReady:        "+CMTI:",
		SMSContent:      "+CMT:",
		SMSStatusReport: "+CDS:",
		CellBroadcast:   "+CBM:",

		// 网络注册相关
		NetworkReg: "+CREG:",
		GPRSReg:    "+CGREG:",
		EPSReg:     "+CEREG:",

		// 服务相关
		USSD: "+CUSD:",

		// 状态变化
		StatusChange: "+CIEV:",

		// 连接相关
		DataCallStatus: "+CGEV:",

		// 时区和网络
		NetworkTime: "+CTZV:",
	}
}

// GetAllNotifications 返回所有URC前缀的列表
func (ns *NotificationSet) GetAllNotifications() []string {
	return []string{
		// 通话相关
		ns.Ring,
		ns.CallRing,
		ns.CallerID,
		ns.CallWaiting,
		ns.ConnectedLine,
		ns.SuppressNotify,
		ns.UnsolicitedNotify,

		// 短信相关
		ns.SMSReady,
		ns.SMSContent,
		ns.SMSStatusReport,
		ns.CellBroadcast,

		// 网络注册相关
		ns.NetworkReg,
		ns.GPRSReg,
		ns.EPSReg,

		// 服务相关
		ns.USSD,

		// 状态变化
		ns.StatusChange,

		// 连接相关
		ns.DataCallStatus,

		// 时区和网络
		ns.NetworkTime,
	}
}

// IsNotification 检查给定行是否为URC
func (ns *NotificationSet) IsNotification(line, cmd string) bool {
	urc := ""
	for _, item := range ns.GetAllNotifications() {
		if item != "" && strings.HasPrefix(line, item) {
			urc = item
			break
		}
	}
	if cmd != "" && urc != "" {
		urc = "AT" + strings.TrimSuffix(urc, ":")
		return !strings.HasPrefix(cmd, urc)
	}
	return urc != ""
}
