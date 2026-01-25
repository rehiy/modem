# Modem

Modem 是一个用于与 GSM/4G/5G 调制解调器设备通信的 Go 语言库集合。

## 模块

### at - AT 命令通信库

提供完整的 AT 命令接口，用于与调制解调器进行通信。

**文档:** [at/README.md](at/README.md)

**主要功能:**

- AT 命令发送和响应处理
- 智能响应和通知识别
- 设备信息查询
- 信号质量监控
- 网络状态管理
- 通话功能
- 短信发送和接收

**快速使用:**

```go
import "github.com/rehiy/modem/at"

device := at.New(port, urcHandler, config)
responses, err := device.SendCommand("AT+CREG?")
```

### dev - 设备预设配置

提供特定 Modem 设备的预设配置，简化设备初始化。

**当前支持设备:**

- ML307A - 中移物联模块

**快速使用:**

```go
import "github.com/rehiy/modem/dev"

// 获取 ML307A 设备预设配置
ml307a := dev.NewML307A()
config := &at.Config{
    CommandSet:      ml307a.CommandSet,
    ResponseSet:     ml307a.ResponseSet,
    NotificationSet: ml307a.NotificationSet,
}
device := at.New(port, urcHandler, config)
```

### sms - 短信编码/解码库

提供 SMS TPDU 的编码和解码功能，遵循 3GPP 规范。

**文档:** [sms/README.md](sms/README.md)

**主要功能:**

- SMS TPDU 编码和解码
- 长消息自动分段
- 自动字符集选择（GSM 7-bit / UCS2）
- 支持中文和表情符号
- 消息收集和重组

**快速使用:**

```go
import "github.com/rehiy/modem/sms"

// 编码
tpdus, _ := sms.Encode([]byte("hello world"), sms.To("+8613800138000"))

// 解码
msg, _ := sms.Decode(tpdus)
```

## 依赖

- Go 1.21+（使用 `atomic.Bool`、`slices.Clone` 等特性）

## 许可证

MIT License
