# Modem AT Command Library

Go语言AT命令库，用于与各种调制解调器进行通信。

## 项目概述

本项目提供了一个功能完整的Go语言库，用于通过AT命令与调制解调器设备进行通信。支持异步命令处理、短信收发、实时事件监听等功能。

## 子模块

- [`at/`](./at/) - 核心AT命令包，提供完整的AT命令处理功能

## 功能特性

- 🔄 **异步处理** - 支持并发安全的命令执行
- 📱 **短信支持** - 完整的短信收发和管理功能
- 📡 **实时监听** - 异步事件监听和处理
- ⚡ **高性能** - 优化的串口通信和命令序列化
- 🛡️ **错误处理** - 详细的错误分类和处理机制
- 🔧 **灵活配置** - 丰富的配置选项和参数设置

## 快速开始

### 安装

```bash
go get github.com/rehiy/modem
```

### 基础使用

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/rehiy/modem/at"
)

func main() {
    // 创建AT连接
    modem, err := at.New(&at.Options{
        Port:     "/dev/ttyUSB0",
        Baudrate: 115200,
        Timeout:  5 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer modem.Close()
    
    // 发送AT命令
    resp, err := modem.Command("ATI", 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Modem Info:", resp)
}
```

## 文档

详细的API文档和使用示例请参考：

- [AT命令包详细文档](./at/README.md)

## 支持的设备

本库支持大多数标准的AT命令调制解调器设备，包括：

- **4G/5G 调制解调器** - 支持各种品牌和型号
- **3G 调制解调器** - 兼容传统3G设备
- **GSM 调制解调器** - 支持标准GSM模块
- **串口设备** - 任何支持AT命令的串口设备

## 系统要求

- Go 1.18+
- Linux/Windows/macOS
- 串口设备访问权限
