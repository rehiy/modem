# SMS TPDU 编码/解码库

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

一个功能完善的 Go 语言 SMS TPDU 编码/解码库，严格遵循 3GPP TS 23.040 和 3GPP TS 23.038 规范。

## 目录

- [简介](#简介)
- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [核心概念](#核心概念)
- [使用指南](#使用指南)
- [高级选项](#高级选项)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)
- [规范参考](#规范参考)

## 简介

SMS 库提供 SMS TPDU（Transport Protocol Data Unit）的编码和解码功能，最初设计用于通过 GSM 调制解调器发送和接收短信，但也可用于任何需要处理 SMS TPDU 或其字段的场景。

### 致谢

本库基于 [warthog618/sms](https://github.com/warthog618/sms) 进行修改。由于原贡献者已不再维护该包，而项目需要根据实际需求进行一些细节调整，因此复制并修改了原始代码。感谢原作者做出的杰出贡献。

## 功能特性

### 核心功能

- ✅ 从 UTF-8 字符串创建 SMS TPDU，包括表情符号 😁
- ✅ 将长消息自动分割为多个连续的 SMS TPDU
- ✅ 编码时自动选择字符集和语言
- ✅ 将 SMS TPDU 解码为 UTF-8 字符串
- ✅ 将连续的 SMS TPDU 重新组合为完整的长消息
- ✅ 支持 PDU 模式的 SMS TPDU 编码和解码

### 字符集支持

- 🌍 完整的 GSM 字符集支持（GSM 7-bit、8-bit、UCS2）
- 🌐 支持多种语言和特殊字符
- 📝 自动字符集检测和选择

## 快速开始

### 安装

```bash
go get github.com/rehiy/modem
```

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/sms"
)

func main() {
    // 1. 编码消息
    tpdus, _ := sms.Encode([]byte("hello world"))
    for _, p := range tpdus {
        b, _ := p.MarshalBinary()
        fmt.Printf("PDU: %X\n", b)
    }

    // 2. 解码消息
    msg, _ := sms.Decode(tpdus)
    fmt.Printf("解码消息: %s\n", msg)
}
```

## 核心概念

### TPDU 类型

SMS TPDU 分为两种主要类型：

| 类型 | 方向 | 说明 |
|------|------|------|
| SMS-SUBMIT | MO (Mobile Originated) | 从移动设备发送到网络 |
| SMS-DELIVER | MT (Mobile Terminated) | 从网络发送到移动设备 |

### 长消息分段

当消息超过单条短信限制时：

- GSM 7-bit 编码：153 字符/段
- UCS2 编码：67 字符/段
- 自动使用 UDH（User Data Header）管理分段信息

## 使用指南

### 编码 (Encoding)

#### 单条消息编码

```go
import "github.com/rehiy/modem/sms"

tpdus, _ := sms.Encode([]byte("hello world"))
for _, p := range tpdus {
    b, _ := p.MarshalBinary()
    // 发送二进制 TPDU 到调制解调器...
}
```

#### 带目标号码的编码

```go
// 指定接收号码
tpdus, _ := sms.Encode([]byte("hello"), sms.To("+8613800138000"))
```

#### 多条消息编码

发送多条消息需要维护计数器，使用 `sms.Encoder`：

```go
e := sms.NewEncoder()
for msg := range msgChan {
    tpdus, _ := e.Encode(msg)
    for _, p := range tpdus {
        b, _ := p.MarshalBinary()
        // 发送二进制 TPDU...
    }
}
```

### 反序列化 (Unmarshalling)

将接收到的二进制 TPDU 转换为 TPDU 对象：

```go
pdu, _ := sms.Unmarshal(bintpdu)
```

### 解码 (Decoding)

#### 单条消息解码

```go
msg, _ := sms.Decode([]*tpdu.TPDU{pdu})
fmt.Println(msg)
```

#### 长消息解码

```go
// tpdus 是包含同一消息多个分段的数组
msg, _ := sms.Decode(tpdus)
fmt.Println(msg)
```

### 消息收集 (Collection)

长消息的分段必须在解码前收集。使用 `Collector` 收集分段：

```go
c := sms.NewCollector()
for {
    bintpdu := <- pduChan
    pdu, _ := sms.Unmarshal(bintpdu)
    
    // 收集分段
    tpdus, _ := c.Collect(pdu)
    
    // 当接收到完整消息的所有分段时
    if len(tpdus) > 0 {
        msg, _ := sms.Decode(tpdus)
        // 处理完整消息...
    }
}
```

### 完整示例

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/sms"
)

func main() {
    // 示例 1: 编码并发送
    message := "Hello, 这是一条长消息，需要分段发送"
    tpdus, _ := sms.Encode([]byte(message), sms.To("+8613800138000"))
    
    for i, p := range tpdus {
        b, _ := p.MarshalBinary()
        fmt.Printf("分段 %d: %X\n", i+1, b)
    }

    // 示例 2: 接收并解码
    // 假设从调制解调器接收到多个 TPDU
    collector := sms.NewCollector()
    
    for _, bintpdu := range receivedTpdus {
        pdu, _ := sms.Unmarshal(bintpdu)
        coll, _ := collector.Collect(pdu)
        
        if len(coll) > 0 {
            msg, _ := sms.Decode(coll)
            fmt.Printf("完整消息: %s\n", msg)
        }
    }
}
```

## 高级选项

### 编码选项

#### 指定目标号码

```go
tpdus, _ := sms.Encode("hello", sms.To("+8613800138000"))
```

#### 指定发送号码（SMS-DELIVER）

```go
tpdus, _ := sms.Encode("hello", sms.From("+8613800138000"), sms.AsDeliver)
```

#### 使用特定字符集

```go
import "github.com/rehiy/modem/sms/charset"

// 使用 Urdu 字符集
tpdus, _ := sms.Encode("hello ٻ", sms.WithCharset(charset.Urdu))
```

#### 强制编码方式

```go
// 强制使用 UCS2 编码
tpdus, _ := sms.Encode("你好", sms.AsUCS2)

// 强制使用 8-bit 编码
tpdus, _ := sms.Encode(data, sms.As8Bit)
```

### 解码选项

#### 限制字符集

```go
// 仅使用默认字符集
msg, _ := sms.Decode(tpdus, sms.WithDefaultCharset())

// 使用指定字符集
msg, _ := sms.Decode(tpdus, sms.WithCharset(charset.Turkish))
```

### 反序列化选项

#### 指定 TPDU 方向

```go
// 从移动台发起
pdu, _ := sms.Unmarshal(bintpdu, sms.AsMO)

// 在移动台终止（默认）
pdu, _ := sms.Unmarshal(bintpdu, sms.AsMT)
```

### 完整选项列表

| 选项 | 类别 | 描述 |
|------|------|------|
| `WithReassemblyTimeout(duration,handler)` | Collect | 限制等待完整重新组装的 TPDU 的时间 |
| `WithTemplate(tpdu)` | Encode | 使用提供的 TPDU 作为编码 TPDU 的模板 |
| `WithTemplateOption(tpdu.Option)` | Encode | 在编码期间将提供的选项应用于模板 TPDU |
| `To(number)` | Encode | 将编码 TPDU 的 DA（目的地址）设置为提供的号码 |
| `From(number)` | Encode | 将编码 TPDU 的 OA（源地址）设置为提供的号码 |
| `WithAllCharsets` | Decode,Encode | 使所有 GSM7 字符集可用 |
| `WithDefaultCharset` | Decode,Encode | 仅使默认字符集可用 |
| `WithCharset(nli...)` | Decode,Encode | 使指定的字符集可用 |
| `WithLockingCharset(nli...)` | Decode,Encode | 使指定的字符集可作为锁定字符集使用 |
| `WithShiftCharset(nli...)` | Decode,Encode | 使指定的字符集可作为移位字符集使用 |
| `AsSubmit` | Encode | 将 TPDU 编码为 SMS-SUBMIT（默认） |
| `AsDeliver` | Encode | 将 TPDU 编码为 SMS-DELIVER |
| `As8Bit` | Encode | 强制将用户数据编码为 8 位 |
| `AsUCS2` | Encode | 强制将用户数据编码为 UCS-2 |
| `AsMO` | Unmarshal | 将 TPDU 视为从移动台发起 |
| `AsMT` | Unmarshal | 将 TPDU 视为在移动台终止（默认） |

## 最佳实践

### 1. 字符集处理

```go
// 自动选择：让库自动选择最佳字符集
tpdus, _ := sms.Encode("Hello 你好")

// 手动指定：当需要兼容特定设备时
tpdus, _ := sms.Encode("Hello 你好", sms.AsUCS2)
```

### 2. 长消息处理

```go
// 自动分段：库会自动处理长消息
longMsg := "这是一条很长的消息，会被自动分成多个分段..."
tpdus, _ := sms.Encode([]byte(longMsg), sms.To("+8613800138000"))

// 分段数查询
fmt.Printf("消息分为 %d 段\n", len(tpdus))
```

### 3. 消息收集超时

```go
// 设置分段收集超时
c := sms.NewCollector(
    sms.WithReassemblyTimeout(5*time.Minute, func(reference, total, received uint8) {
        fmt.Printf("分段收集超时: ref=%d, total=%d, received=%d\n", reference, total, received)
    }),
)
```

### 4. 错误处理

```go
tpdus, err := sms.Encode([]byte("hello"))
if err != nil {
    // 处理编码错误
    log.Printf("编码失败: %v", err)
    return
}

for _, p := range tpdus {
    b, err := p.MarshalBinary()
    if err != nil {
        log.Printf("序列化失败: %v", err)
        continue
    }
    // 发送 PDU...
}
```

## 常见问题

### Q1: 如何发送中文短信？

```go
// 方法 1: 自动处理（推荐）
tpdus, _ := sms.Encode("你好世界", sms.To("+8613800138000"))

// 方法 2: 强制 UCS2 编码
tpdus, _ := sms.Encode("你好世界", sms.To("+8613800138000"), sms.AsUCS2)
```

### Q2: 如何发送带表情符号的短信？

```go
tpdus, _ := sms.Encode("Hello 😁 World", sms.To("+8613800138000"), sms.AsUCS2)
```

### Q3: 一条消息最多能有多长？

| 编码方式 | 单条长度 | 分段长度 |
|---------|---------|---------|
| GSM 7-bit | 160 字符 | 153 字符/段 |
| UCS2 | 70 字符 | 67 字符/段 |

### Q4: 如何处理接收到的多分段消息？

```go
collector := sms.NewCollector()
for {
    pduBytes := receiveFromModem()
    pdu, _ := sms.Unmarshal(pduBytes)
    
    segments, _ := collector.Collect(pdu)
    if len(segments) > 0 {
        // 收集到完整的消息
        msg, _ := sms.Decode(segments)
        processMessage(msg)
    }
}
```

### Q5: 如何调试 TPDU？

```go
tpdus, _ := sms.Encode("hello")
for i, p := range tpdus {
    b, _ := p.MarshalBinary()
    fmt.Printf("分段 %d: %X\n", i+1, b)
    
    // 查看 TPDU 详细信息
    fmt.Printf("类型: %T\n", p)
    fmt.Printf("长度: %d\n", p.Length())
}
```

## 规范参考

本库严格遵循以下规范：

- **3GPP TS 23.040** - Technical realization of the Short Message Service (SMS)
- **3GPP TS 23.038** - Alphabets and language-specific information

这些规范定义了 SMS 的技术实现细节，包括 TPDU 格式、编码规则、字符集等。

## 许可证

MIT License
