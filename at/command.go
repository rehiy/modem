package at

// CommandSet 定义可配置的 AT 命令集
type CommandSet struct {
	// 基本控制命令
	Test         string // 测试连接 AT
	EchoOff      string // 关闭回显 ATE0
	EchoOn       string // 开启回显 ATE1
	Reset        string // 重置 modem ATZ
	FactoryReset string // 恢复出厂设置 AT&F
	SaveSettings string // 保存设置 AT&W
	LoadProfile  string // 加载配置文件 ATZ[profile]
	SaveProfile  string // 保存到配置文件 AT&W[profile]

	// 设备身份信息
	IMEI         string // 查询 IMEI AT+CGSN
	Manufacturer string // 查询制造商 AT+CGMI
	Model        string // 查询型号 AT+CGMM
	Revision     string // 查询版本 AT+CGMR
	IMSI         string // 查询 IMSI AT+CIMI
	ICCID        string // 查询 ICCID AT+CCID
	Number       string // 查询本机号码 AT+CNUM

	// 网络状态
	Operator    string // 查询/设置运营商、网络模式 AT+COPS
	NetworkMode string // 查询/设置网络模式 AT+CNMP
	NetworkReg  string // 查询/设置网络注册状态及通知 AT+CREG
	GPRSReg     string // 查询/设置 GPRS 注册状态及通知 AT+CGREG
	Signal      string // 查询信号质量/设置上报 AT+CSQ

	// SIM 卡管理
	SIMStatus string // 查询/验证 SIM 卡状态 AT+CPIN
	PINVerify string // 验证 PIN 码 AT+CPIN
	PINChange string // 修改 PIN 码 AT+CPWD
	PINLock   string // 查询/设置 PIN 锁状态 AT+CLCK

	// 设备状态
	BatteryLevel string // 查询电池电量 AT+CBC
	DeviceTemp   string // 查询设备温度 AT+CPMUTEMP
	NetworkTime  string // 查询/设置网络时间 AT+CCLK
	SetTime      string // 设置时间 AT+CCLK

	// 网络配置
	APN        string // 查询/设置 APN 配置 AT+CGDCONT
	IPAddress  string // 查询 IP 地址 AT+CGPADDR
	PDPContext string // 查询/设置 PDP 上下文状态 AT+CGACT
	SetAPN     string // 设置 APN AT+CGDCONT

	// 短信相关
	SmsFormat string // 查询/设置短信格式 AT+CMGF
	SmsStore  string // 查询/设置短信存储位置 AT+CPMS
	SmsCenter string // 查询/设置短信中心号码 AT+CSCA
	ListSms   string // 列出短信 AT+CMGL
	ReadSms   string // 读取短信 AT+CMGR
	DeleteSms string // 删除短信 AT+CMGD
	SendSms   string // 发送短信 AT+CMGS

	// 语音通话
	Dial      string // 拨号 ATD
	Answer    string // 接听 ATA
	Hangup    string // 挂断 ATH
	CallerID  string // 查询/设置来电显示 AT+CLIP
	CallState string // 查询通话状态 AT+CLCC
	CallWait  string // 查询/设置呼叫等待 AT+CCWA
	CallFWD   string // 查询/设置呼叫转移 AT+CCFC

	// 通知管理
	NetworkRegNotify string // 查询/设置网络注册通知 AT+CREG
	GPRSRegNotify    string // 查询/设置 GPRS 注册通知 AT+CGREG
	SignalReport     string // 设置信号质量上报 AT+CSQ
}

// DefaultCommandSet 返回默认的标准 AT 命令集
func DefaultCommandSet() *CommandSet {
	return &CommandSet{
		// 基本控制命令
		Test:         "AT",
		EchoOff:      "ATE0",
		EchoOn:       "ATE1",
		Reset:        "ATZ",
		FactoryReset: "AT&F",
		SaveSettings: "AT&W",
		LoadProfile:  "ATZ",
		SaveProfile:  "AT&W",

		// 设备身份信息
		IMEI:         "AT+CGSN",
		Manufacturer: "AT+CGMI",
		Model:        "AT+CGMM",
		Revision:     "AT+CGMR",
		IMSI:         "AT+CIMI",
		ICCID:        "AT+CCID",
		Number:       "AT+CNUM",

		// 网络状态
		Operator:    "AT+COPS",
		NetworkMode: "AT+CNMP",
		NetworkReg:  "AT+CREG",
		GPRSReg:     "AT+CGREG",
		Signal:      "AT+CSQ",

		// SIM 卡管理
		SIMStatus: "AT+CPIN",
		PINVerify: "AT+CPIN",
		PINChange: "AT+CPWD",
		PINLock:   "AT+CLCK",

		// 设备状态
		BatteryLevel: "AT+CBC",
		DeviceTemp:   "AT+CPMUTEMP",
		NetworkTime:  "AT+CCLK",
		SetTime:      "AT+CCLK",

		// 网络配置
		APN:        "AT+CGDCONT",
		IPAddress:  "AT+CGPADDR",
		PDPContext: "AT+CGACT",
		SetAPN:     "AT+CGDCONT",

		// 短信相关
		SmsFormat: "AT+CMGF",
		SmsStore:  "AT+CPMS",
		SmsCenter: "AT+CSCA",
		ListSms:   "AT+CMGL",
		ReadSms:   "AT+CMGR",
		DeleteSms: "AT+CMGD",
		SendSms:   "AT+CMGS",

		// 语音通话
		Dial:      "ATD",
		Answer:    "ATA",
		Hangup:    "ATH",
		CallerID:  "AT+CLIP",
		CallState: "AT+CLCC",
		CallWait:  "AT+CCWA",
		CallFWD:   "AT+CCFC",

		// 通知管理
		NetworkRegNotify: "AT+CREG",
		GPRSRegNotify:    "AT+CGREG",
		SignalReport:     "AT+CSQ",
	}
}
