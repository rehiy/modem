# AT Command Package

一个用于管理调制解调器的Go语言AT命令包，提供简洁的API来发送AT命令和处理调制解调器响应。

## 功能特性

- 🚀 **异步处理**：使用goroutine实现异步命令执行和响应处理
- 🔧 **模块化设计**：清晰的文件结构，职责分离明确
- 📡 **命令管理**：支持标准AT命令和SMS命令
- 🔔 **异步指示**：支持处理调制解调器的异步通知
- ⚡ **超时控制**：可配置的命令超时机制
- 🛡️ **连接管理**：自动检测连接状态和错误处理

## 文件结构

```text
at/
├── at.go          # 核心结构体和构造函数
├── cmd.go         # 命令执行相关功能
├── indication.go  # 异步指示处理
├── loop.go        # 内部循环处理
├── options.go     # 配置选项系统
├── error.go       # 错误类型定义
├── parser.go      # 命令解析辅助函数
└── README.md      # 本文档
```

## 安装使用

```go
go get github.com/rehiy/modem
```

## 快速开始

```go
package main

import (
    "fmt"
    "io"
    "github.com/rehiy/modem/at"
)

func main() {
    // 假设modem是实现了io.ReadWriter接口的设备
    var modem io.ReadWriter
    
    // 创建AT实例并指定初始化命令
    atModem := at.New(modem, at.WithCmds("Z", "E0", "+CMEE=1"))
    
    // 发送AT命令
    response, err := atModem.Command("AT+CSQ")
    if err != nil {
        fmt.Printf("命令执行错误: %v\n", err)
        return
    }
    
    fmt.Printf("信号强度: %s\n", response)
}
```

## API 文档

### 核心类型

#### AT 结构体

```go
type AT struct {
    // 内部字段，通过方法访问
}
```

#### 主要方法

- `New(modem io.ReadWriter, options ...Option) *AT` - 创建新的AT实例
- `Command(cmd string, options ...CommandOption) ([]string, error)` - 执行AT命令
- `SMSCommand(cmd string, sms string, options ...CommandOption) ([]string, error)` - 执行SMS相关命令
- `AddIndication(prefix string, handler InfoHandler, options ...IndicationOption) error` - 添加异步指示处理器
- `CancelIndication(prefix string)` - 移除指示处理器
- `Closed() <-chan struct{}` - 获取连接状态通道

### 配置选项

#### 构造函数选项

- `WithEscTime(d time.Duration) EscTimeOption` - 设置转义保护时间（默认：20ms）
- `WithCmds(cmds ...string) CmdsOption` - 设置初始化命令（默认：ATZ, ATE0）
- `WithTimeout(d time.Duration) TimeoutOption` - 设置命令超时（默认：1s）

#### 命令选项

- `WithTimeout(d time.Duration) TimeoutOption` - 设置单个命令超时

#### 指示选项

- `WithTrailingLines(l int) TrailingLinesOption` - 设置指示的尾随行数
- `WithTrailingLine` - 预定义的一个尾随行选项

### 错误类型

- `ErrClosed` - 操作无法执行，调制解调器已关闭
- `ErrDeadlineExceeded` - 调制解调器未在要求时间内完成操作
- `ErrError` - 调制解调器返回通用AT ERROR
- `ErrIndicationExists` - 前缀的指示已注册
- `CMEError` - 调制解调器返回的CME错误
- `CMSError` - 调制解调器返回的CMS错误
- `ConnectError` - 拨号尝试失败

### 类型定义

- `InfoHandler func([]string)` - 指示信息的处理器函数
- `IndicationOption` - 指示配置选项的接口

### API 参考表格

| 类别 | 方法/类型 | 描述 | 参数 | 返回值 |
|------|-----------|------|------|--------|
| **构造函数** | `New` | 创建AT实例 | `modem io.ReadWriter`, `options ...Option` | `*AT` |
| **核心方法** | `Command` | 执行AT命令 | `cmd string`, `options ...CommandOption` | `[]string, error` |
| | `SMSCommand` | 执行SMS命令 | `cmd string`, `sms string`, `options ...CommandOption` | `[]string, error` |
| | `AddIndication` | 注册指示处理器 | `prefix string`, `handler InfoHandler`, `options ...IndicationOption` | `error` |
| | `CancelIndication` | 移除指示处理器 | `prefix string` | - |
| | `Closed` | 获取连接状态 | - | `<-chan struct{}` |
| **选项** | `WithEscTime` | 设置转义保护时间 | `d time.Duration` | `EscTimeOption` |
| | `WithCmds` | 设置初始化命令 | `cmds ...string` | `CmdsOption` |
| | `WithTimeout` | 设置超时 | `d time.Duration` | `TimeoutOption` |
| | `WithTrailingLines` | 设置尾随行数 | `l int` | `TrailingLinesOption` |
| **错误** | `ErrClosed` | 连接关闭错误 | - | `error` |
| | `ErrDeadlineExceeded` | 超时错误 | - | `error` |
| | `CMEError` | CME错误类型 | - | `error` |
| | `CMSError` | CMS错误类型 | - | `error` |

### 常用AT命令参考

| 命令 | 描述 | 使用示例 |
|------|------|----------|
| `ATI` | 制造商识别 | `atModem.Command("I")` |
| `AT+CSQ` | 信号质量 | `atModem.Command("+CSQ")` |
| `AT+CGMI` | 制造商信息 | `atModem.Command("+CGMI")` |
| `AT+CGMM` | 型号信息 | `atModem.Command("+CGMM")` |
| `AT+CMGF=1` | 设置文本模式 | `atModem.Command("+CMGF=1")` |
| `AT+CMGS` | 发送短信 | `atModem.SMSCommand("+CMGS=\"手机号\"", "消息")` |
| `AT+CNUM` | 本机号码 | `atModem.Command("+CNUM")` |

**注意**：命令不应包含"AT"前缀或行结束符 - 这些由包自动处理。

### 详细 API 使用

#### Command 方法

```go
// 执行标准AT命令
response, err := atModem.Command("AT+CSQ")
if err != nil {
    log.Printf("信号查询失败: %v", err)
} else {
    log.Printf("信号质量响应: %v", response)
}

// 使用超时选项执行命令
response, err := atModem.Command("AT+CGMI", at.WithTimeout(30*time.Second))
if err != nil {
    log.Printf("制造商查询失败: %v", err)
} else {
    log.Printf("制造商: %v", response)
}
```

#### SMSCommand 方法

```go
// 在文本模式下执行SMS命令
response, err := atModem.SMSCommand("+CMGS=\"+1234567890\"", "你好世界！")
if err != nil {
    log.Printf("短信发送失败: %v", err)
} else {
    log.Printf("短信发送成功: %v", response)
}

// 执行带超时的SMS命令
response, err := atModem.SMSCommand("+CMGS=\"+1234567890\"", "测试消息", 
    at.WithTimeout(60*time.Second))
```

#### 指示处理

```go
// 定义指示处理器
smsHandler := func(lines []string) {
    log.Printf("收到短信指示: %v", lines)
    // 从lines中处理短信内容
}

// 注册短信接收指示
err := atModem.AddIndication("+CMT:", smsHandler, at.WithTrailingLines(1))
if err != nil {
    log.Printf("指示注册失败: %v", err)
    return
}

// 不再需要时取消指示
atModem.CancelIndication("+CMT:")
```

#### 连接状态监控

```go
// 监控连接状态
go func() {
    select {
    case <-atModem.Closed():
        log.Printf("调制解调器连接已关闭")
        // 执行清理操作
    }
}()
```

#### 错误处理示例

```go
// 处理特定错误类型
response, err := atModem.Command("AT+INVALID")
if err != nil {
    switch err {
    case at.ErrDeadlineExceeded:
        log.Printf("命令超时 - 调制解调器未响应")
    case at.ErrClosed:
        log.Printf("连接已关闭 - 调制解调器已断开")
    case at.ErrError:
        log.Printf("调制解调器返回ERROR - 无效命令")
    case at.CMEError:
        log.Printf("CME错误: %v", err)
    case at.CMSError:
        log.Printf("CMS错误: %v", err)
    default:
        log.Printf("命令执行错误: %v", err)
    }
}
```

### 高级用法

#### 自定义初始化命令

```go
// 使用自定义初始化创建AT实例
atModem := at.New(modem,
    at.WithCmds("Z", "E0", "+CMEE=1"),  // 重置、禁用回显、启用详细错误
    at.WithEscTime(50*time.Millisecond),  // 设置更长的转义时间
    at.WithTimeout(5*time.Second),         // 设置默认命令超时
)
```

#### 多个指示处理器

```go
// 注册多个指示
callHandler := func(lines []string) {
    log.Printf("来电: %v", lines)
}

networkHandler := func(lines []string) {
    log.Printf("网络状态变化: %v", lines)
}

atModem.AddIndication("RING", callHandler)
atModem.AddIndication("+CREG:", networkHandler, at.WithTrailingLine)
```

### 参数说明

- **命令字符串**: 不应包含"AT"前缀或"\r\n"后缀（自动添加）
- **SMS命令**: 两步过程 - 命令行后跟SMS数据
- **超时值**: 使用`time.Duration`（例如：`30*time.Second`）
- **指示前缀**: 精确匹配调制解调器响应前缀
- **处理器函数**: 接收包含指示行的字符串切片

这份全面的API文档应该帮助用户有效地利用AT命令包进行调制解调器通信任务。

### 设计理念

#### 并发安全

所有对调制解调器的访问都通过通道进行序列化，确保并发安全。

#### 模块分离

- **at.go**: 核心协调和生命周期管理
- **cmd.go**: 命令执行和响应处理
- **indication.go**: 异步通知处理
- **parser.go**: 响应解析逻辑
- **options.go**: 配置选项系统

#### 错误处理

提供清晰的错误类型和详细的错误信息，便于调试和问题排查。

### 设计哲学

#### 并发安全性

所有调制解调器访问都通过通道进行序列化，确保并发安全。

### 贡献指南

欢迎提交Issue和Pull Request来改进这个包。

### 许可证

[MIT License](LICENSE)

---

## Acknowledgments

This project is based on the AT command package implementation from [warthog618/modem](https://github.com/warthog618/modem/blob/master/at/at.go).

Special thanks to the original author for the excellent design and implementation, which provided us with a stable and reliable foundation for AT command processing. We have performed modular refactoring and functional enhancements on this basis, but the core design philosophy and architectural ideas all originate from the original project.

**Salute to the original author's open source contribution!** 🙏
