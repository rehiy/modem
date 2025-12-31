# Go AT 命令通信库

一个轻量级的 Go 语言 AT 命令通信库，用于和串口 Modem 设备进行交互。

## 功能特性

- **完整的 AT 命令接口** - 基础命令、信息查询、信号质量、网络状态、通话、短信等
- **智能响应处理** - 自动识别最终响应（OK/ERROR 等）和通知消息（URC）
- **并发安全** - 使用原子操作和互斥锁保证线程安全
- **可扩展配置** - 支持自定义命令集、响应集和通知集
- **短信功能** - 自动编码检测（ASCII/UCS2）、长短信自动分段
- **通知监听** - 来电、短信、网络状态变化等实时通知

## 安装

```bash
go get github.com/rehiy/modem
```

## 快速开始

### 基本使用

```go
package main

import (
 "log"
 "time"

 "github.com/rehiy/modem/at"
)

func main() {
 // 创建串口连接（需自行实现 Port 接口）
 port := openSerialPort("/dev/ttyUSB0", 115200)
 defer port.Close()

 // 配置通知处理函数
 urcHandler := func(label string, param map[int]string) {
  log.Printf("通知: %s %v", label, param)
 }

 // 创建设备
 config := &at.Config{
  Timeout: 5 * time.Second,
 }
 device := at.New(port, urcHandler, config)
 defer device.Close()

 // 测试连接
 if err := device.Test(); err != nil {
  log.Fatal(err)
 }

 // 查询设备信息
 manufacturer, _ := device.GetManufacturer()
 model, _ := device.GetModel()
 log.Printf("设备: %s %s", manufacturer, model)
}
```

## 核心接口

### Port 接口

`Port` 接口定义了与串口设备交互的基本方法：

```go
type Port interface {
 Read(buf []byte) (int, error)
 Write(data []byte) (int, error)
 Flush() error
 Close() error
}
```

### Device 方法

```go
// 创建设备连接
func New(port Port, handler UrcHandler, config *Config) *Device

// 连接管理
func (m *Device) IsOpen() bool
func (m *Device) Close() error

// 命令发送
func (m *Device) SendCommand(cmd string) ([]string, error)
func (m *Device) SendCommandExpect(cmd, expected string) error
```

## 配置说明

### Config 结构

```go
type Config struct {
 Timeout         time.Duration        // 超时时间（默认 1秒）
 CommandSet      *CommandSet          // 自定义 AT 命令集
 ResponseSet     *ResponseSet         // 自定义响应类型集
 NotificationSet *NotificationSet     // 自定义通知类型集
 Printf          func(string, ...any) // 日志输出函数，如果为 nil 则使用 log.Printf
}
```

### CommandSet 结构

定义可配置的 AT 命令集：

```go
type CommandSet struct {
 // 基本命令
 Test, EchoOff, EchoOn, Reset, FactoryReset, SaveSettings string
 // 信息查询
 Manufacturer, Model, Revision, SerialNumber, IMSI, ICCID, PhoneNumber, Operator string
 // 信号质量
 SignalQuality string
 // 网络注册
 NetworkRegistration, GPRSRegistration string
 // 短信相关
 SMSFormat, ListSMS, ReadSMS, DeleteSMS, SendSMS string
 // 通话相关
 Dial, Answer, Hangup, CallerID string
}
```

使用 `DefaultCommandSet()` 获取标准 AT 命令集。

### ResponseSet 结构

定义命令响应类型集合：

```go
type ResponseSet struct {
 OK, Error, NoCarrier, NoAnswer, NoDialtone, Busy, Connect string
 CMEError, CMSError string
 CustomFinal []string // 自定义最终响应
}
```

使用 `DefaultResponseSet()` 获取默认响应集。

### NotificationSet 结构

定义 URC（Unsolicited Result Code）通知类型集合：

```go
type NotificationSet struct {
 Ring, SMSReady, SMSContent, SMSStatusReport, CellBroadcast string
 CallRing, CallerID, CallWaiting string
 NetworkReg, GPRSReg, EPSReg, USSD, StatusChange string
}
```

使用 `DefaultNotificationSet()` 获取默认通知集。

## 设备命令

### 基本命令

```go
// 测试连接
device.Test()

// 回显控制
device.EchoOff()
device.EchoOn()

// 重置和保存
device.Reset()
device.FactoryReset()
device.SaveSettings()
```

### 信息查询

```go
manufacturer, _ := device.GetManufacturer()
model, _ := device.GetModel()
revision, _ := device.GetRevision()
serial, _ := device.GetSerialNumber()
imsi, _ := device.GetIMSI()
iccid, _ := device.GetICCID()
phoneNumber, _ := device.GetPhoneNumber()
// 查询运营商信息
// 返回 (mode, format, operator, act, error)
// mode: 网络选择模式 0-4
// format: 格式编号
// operator: 运营商（如 "C46001"）
// act: 无线接入技术类型
mode, operator, format, _ := device.GetOperator()
```

### 信号和网络

```go
// 信号质量：返回 (rssi, ber, error)
// rssi: 信号强度 0-31（99 表示未知）
// ber: 误码率 0-7（99 表示未知）
rssi, ber, _ := device.GetSignalQuality()

// 网络注册状态：返回 (n, stat, error)
// n: 禁用/启用状态
// stat: 注册状态 0-5
n, stat, _ := device.GetNetworkStatus()

// GPRS 注册状态
n, stat, _ := device.GetGPRSStatus()
```

### 通话功能

```go
// 拨打电话
device.Dial("+8613800138000")

// 接听和挂断
device.Answer()
device.Hangup()

// 来电显示
enabled, _ := device.GetCallerID()
device.SetCallerID(true)
```

## 短信功能

### 短信发送

```go
// 发送短信（自动处理中文和长短信）
device.SendSMS("+8613800138000", "Hello from Go!")
device.SendSMS("+8613800138000", "你好，这是一条中文短信！")
```

**自动编码处理：**

- 纯 ASCII 字符：直接发送，最大 160 字符
- 包含中文：使用 UCS2 编码，最大 70 字符
- 超长消息：自动分段（英文 153 字符/段，中文 67 字符/段）

### 短信管理

```go
// 列出短信
list, _ := device.ListSMS()
for _, sms := range list {
    fmt.Printf("来自: %s\n内容: %s\n", sms.PhoneNumber, sms.Message)
}

// 删除短信
device.DeleteSMS(1) // 删除指定索引
```

### SMS 结构

```go
type SMS struct {
 Index       int    // 短信索引
 Status      string // 状态：REC UNREAD, REC READ, STO UNSENT, STO SENT
 PhoneNumber string // 电话号码
 Timestamp   string // 时间戳
 Message     string // 短信内容
}
```

## 通知处理

通知处理函数在创建设备时传入，自动监听各类 URC（Unsolicited Result Code）：

```go
urcHandler := func(label string, param map[int]string) {
 switch label {
 case "+CMTI:": // 新短信通知
  fmt.Println("收到新短信:", param)
 case "RING":   // 来电
  fmt.Println("电话响铃")
 case "+CLIP:": // 来电显示
  fmt.Println("来电号码:", param)
 case "+CREG:": // 网络状态变化
  fmt.Println("网络状态:", param)
 }
}
```

## 自定义适配

### 自定义命令集

```go
commands := at.DefaultCommandSet()
commands.SignalQuality = "AT^HCSQ"  // 华为扩展命令
commands.ICCID = "AT^ICCID?"

config := &at.Config{
 Timeout:    5 * time.Second,
 CommandSet: &commands,
}
```

### 自定义响应集

```go
responses := at.DefaultResponseSet()
responses.CustomFinal = []string{"CUSTOM_OK"}

config := &at.Config{
 ResponseSet: &responses,
}
```

### 自定义通知集

```go
notifications := at.DefaultNotificationSet()
notifications.SignalQuality = "^HCSQ:"
notifications.NetworkReg = "^CREG:"

config := &at.Config{
 NotificationSet: &notifications,
}
```

## 内部机制

### 通信流程

1. **读取循环** (`readLoop`): 持续从串口读取数据
   - 去除空白字符
   - 识别 URC 通知，交由 `urcHandler` 处理
   - 其他数据写入响应通道

2. **命令发送** (`SendCommand`):
   - 清空响应通道，避免收到残留响应
   - 检查命令是否包含结束符，自动添加默认结束符 `\r\n`
   - 加互斥锁（防止并发写）
   - 发送命令并等待最终响应

3. **响应读取** (`readResponse`):
   - 从响应通道读取数据
   - 检测最终响应（OK/ERROR 等）
   - 超时返回错误

### 并发安全

- `closed`: 使用 `atomic.Bool` 保证原子操作
- `wg`: 使用 `sync.WaitGroup` 等待 goroutine 退出
- `mu`: 互斥锁保护命令发送
- 响应通道: 带缓冲的通道（容量 100）

## 依赖

本库不依赖特定的串口库，用户需要自行实现 `Port` 接口。可参考：

- [github.com/tarm/serial](https://github.com/tarm/serial)
- [go.bug.st/serial](https://github.com/bugst/go-serial)

## 许可证

MIT License
