# rehiy-modem-pdu

[![Go Reference](https://pkg.go.dev/badge/github.com/rehiy/modem.svg)](https://pkg.go.dev/github.com/rehiy/modem)
[![Go Report Card](https://goreportcard.com/badge/github.com/rehiy/modem)](https://goreportcard.com/report/github.com/rehiy/modem)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-87%25-brightgreen.svg)](https://github.com/rehiy/modem)

**高性能的 Go 语言 PDU（Protocol Data Unit）短信编码/解码库**，支持 GSM 7-bit、8-bit 和 UCS2 编码，完全实现 3GPP TS 23.040 标准。

## ✨ 特性

- 🚀 **高性能**：使用现代 Go 标准库优化，编码/解码速度快
- 📱 **完整支持**：GSM 7-bit、8-bit、UCS2 编码全支持
- 📨 **长短信**：自动处理长短信分割和组装
- 🔒 **并发安全**：内置并发安全的长短信管理器
- ✅ **标准兼容**：完全符合 3GPP TS 23.040 和 TS 23.038 标准
- 🧪 **测试完善**：单元测试覆盖率 87%，包含基准测试和竞态检测

## 📦 安装

```bash
go get github.com/rehiy/modem
```

## 🚀 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 编码短信
    msg := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        "Hello World!",
        SMSC:        "+8613800138000",
    }

    pdus, err := pdu.Encode(msg)
    if err != nil {
        panic(err)
    }

    for i, p := range pdus {
        fmt.Printf("PDU %d: %s\n", i+1, p.Data)
    }

    // 解码短信
    pduStr := "07911326040000F0040B911346610089F60000208062917314080CC8329BFD06"
    decoded, err := pdu.Decode(pduStr)
    if err != nil {
        panic(err)
    }

    fmt.Printf("From: %s, Text: %s\n", decoded.PhoneNumber, decoded.Text)
}
```

### 长短信处理

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 创建长消息
    longText := "This is a very long message that will be automatically split into multiple parts..."
    
    msg := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        longText,
        SMSC:        "+8613800138000",
    }

    // 自动分割为多个 PDU
    pdus, err := pdu.Encode(msg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Message split into %d parts\n", len(pdus))

    // 使用长短信管理器组装
    manager := pdu.NewConcatManager()
    
    for _, p := range pdus {
        decoded, _ := pdu.Decode(p.Data)
        complete, err := manager.AddMessage(decoded)
        if err != nil {
            panic(err)
        }
        if complete != nil {
            fmt.Printf("Complete message: %s\n", complete.Text)
        }
    }
}
```

### 中文短信

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    msg := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        "你好世界！",
        SMSC:        "+8613800138000",
        Encoding:    pdu.EncodingUCS2, // 指定 UCS2 编码
    }

    pdus, err := pdu.Encode(msg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("PDU: %s\n", pdus[0].Data)
}
```

## 📚 核心功能

### 支持的编码

- **GSM 7-bit**：默认编码，支持基本拉丁字符和扩展字符
- **8-bit**：二进制数据编码
- **UCS2**：Unicode 编码，支持所有语言

### 消息类型

- **SMS-DELIVER**：接收的短信
- **SMS-SUBMIT**：发送的短信
- **SMS-STATUS-REPORT**：状态报告

### 高级特性

- ✅ 自动编码选择（根据文本内容）
- ✅ 长短信自动分割和组装
- ✅ 并发安全的长短信管理器
- ✅ 闪信支持
- ✅ 状态报告请求
- ✅ 有效期设置

## 🧪 测试

```bash
# 运行所有测试
go test -v ./pdu/

# 竞态检测
go test -race ./pdu/

# 测试覆盖率（87%）
go test -cover ./pdu/
```

## 📄 许可证

MIT 许可证
