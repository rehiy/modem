package at

import (
	"reflect"
	"strings"
)

// NotificationSet URC（Unsolicited Result Code）类型集合
type NotificationSet struct {
	// 通话相关
	Ring              string // RING - 来电响铃
	NoCarrier         string // NO CARRIER - 连接丢失/载波丢失（结果码）
	Busy              string // BUSY - 对方忙（结果码）
	NoAnswer          string // NO ANSWER - 对方未接听（结果码）
	NoDialtone        string // NO DIALTONE - 无拨号音（结果码）
	CallRing          string // +CRING - 响铃指示类型
	CallerID          string // +CLIP - 来电显示
	CallList          string // +CLCC - 当前通话列表
	CallWaiting       string // +CCWA - 呼叫等待
	ConnectedLine     string // +COLP - 连接线号呈现
	SuppressNotify    string // +CSSI - 补充服务通知（失败）- 呼叫抑制
	UnsolicitedNotify string // +CSSU - 补充服务通知（成功）- 呼叫抑制解除

	// 短信相关
	SmsReady        string // +CMTI - 新短信到达通知
	SmsContent      string // +CMT - 短信内容推送
	SmsStatusReport string // +CDS - 短信状态报告
	CellBroadcast   string // +CBM - 小区广播消息
	SmsAck          string // +CNMA - 新消息确认

	// 网络注册
	NetworkReg string // +CREG - GSM 网络注册状态
	GPRSReg    string // +CGREG - GPRS 网络注册状态
	EPSReg     string // +CEREG - EPS (4G) 网络注册状态
	Reg5G      string // +C5GREG - 5G 网络注册状态
	VoiceReg   string // +CIREG - 语音网络注册状态

	// 网络状态
	Operator      string // +COPS - 运营商选择/变化
	SignalQuality string // +CSQ - 信号质量
	NetworkTime   string // +CTZV - 网络时间（NITZ）
	Timezone      string // +CTZU - 时区更新

	// 分组交换域
	PacketEvent string // +CGEV - GPRS 事件通知

	// 移动性管理
	MMStatus5G  string // +5GMM - 5G 移动性管理状态
	MMStatusEPS string // +EMM - EPS 移动性管理状态
	MMStatus    string // +GMM - GMM 移动性管理状态

	// 指示器事件
	IndicationEvent string // +CIEV - 移动设备指示器事件

	// 非3GPP标准（厂商特定扩展）
	CallEnded   string // +CDIS - 呼叫结束通知
	CallHeld    string // +CHLD - 呼叫保持/多方通话状态
	CallForward string // +CCFC - 呼叫转接状态
	SmsSent     string // +CMGS - 短信发送成功
	SmsWrite    string // +CMGW - 短信写入存储
	DeviceReady string // +RDY - 设备就绪
	DeviceBoot  string // +BOOT - 设备启动完成

	// TCP/IP 连接（厂商特定扩展）
	IPConnectOpen  string // +CIPOPEN - IP 连接打开
	IPConnectClose string // +CIPCLOSE - IP 连接关闭
	IPDataReceived string // +CIPRXGOT - IP 数据到达
	IPDataSent     string // +CIPSEND - IP 数据发送状态

	// SIM 卡和错误
	SIMStatus string // +CPIN - SIM 卡状态（PIN）
	CMSError  string // +CMS ERROR - 短信服务错误
	CMEError  string // +CME ERROR - 移动台错误

	// 其他服务
	USSD string // +CUSD - 非结构化补充业务数据
}

// DefaultNotificationSet 返回默认的URC类型集合
func DefaultNotificationSet() *NotificationSet {
	return &NotificationSet{
		// 通话相关
		Ring:              "RING",
		CallRing:          "+CRING",
		CallerID:          "+CLIP",
		CallList:          "+CLCC",
		CallWaiting:       "+CCWA",
		ConnectedLine:     "+COLP",
		SuppressNotify:    "+CSSI",
		UnsolicitedNotify: "+CSSU",
		NoCarrier:         "NO CARRIER",
		Busy:              "BUSY",
		NoAnswer:          "NO ANSWER",
		NoDialtone:        "NO DIALTONE",

		// 短信相关
		SmsReady:        "+CMTI",
		SmsContent:      "+CMT",
		SmsStatusReport: "+CDS",
		CellBroadcast:   "+CBM",
		SmsAck:          "+CNMA",

		// 网络注册
		NetworkReg: "+CREG",
		GPRSReg:    "+CGREG",
		EPSReg:     "+CEREG",
		Reg5G:      "+C5GREG",
		VoiceReg:   "+CIREG",

		// 网络状态
		Operator:      "+COPS",
		SignalQuality: "+CSQ",
		NetworkTime:   "+CTZV",
		Timezone:      "+CTZU",

		// 分组交换域
		PacketEvent: "+CGEV",

		// 移动性管理
		MMStatus5G:  "+5GMM",
		MMStatusEPS: "+EMM",
		MMStatus:    "+GMM",

		// 指示器事件
		IndicationEvent: "+CIEV",

		// 非3GPP标准（厂商特定扩展）
		CallEnded:   "+CDIS",
		CallHeld:    "+CHLD",
		CallForward: "+CCFC",
		SmsSent:     "+CMGS",
		SmsWrite:    "+CMGW",
		DeviceReady: "+RDY",
		DeviceBoot:  "+BOOT",

		// TCP/IP 连接（厂商特定扩展）
		IPConnectOpen:  "+CIPOPEN",
		IPConnectClose: "+CIPCLOSE",
		IPDataReceived: "+CIPRXGOT",
		IPDataSent:     "+CIPSEND",

		// SIM 卡和错误
		SIMStatus: "+CPIN",
		CMSError:  "+CMS ERROR",
		CMEError:  "+CME ERROR",

		// 其他服务
		USSD: "+CUSD",
	}
}

// GetAllNotifications 返回所有URC前缀的列表
func (ns *NotificationSet) GetAllNotifications() []string {
	v := reflect.ValueOf(ns).Elem()

	result := []string{}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String {
			value := field.String()
			if value != "" {
				result = append(result, value)
			}
		}
	}
	return result
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
	// 避免将命令响应误认为 URC
	if cmd != "" && urc != "" && urc[0] == '+' {
		if urc == ns.CMEError || urc == ns.CMSError {
			return false
		}
		if strings.HasPrefix(cmd, "AT"+urc) {
			return false
		}
	}
	return urc != ""
}
