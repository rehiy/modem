# SMS

SMS 是一个用于编码和解码 SMS TPDU 的 Go 语言库，遵循 3GPP TS 23.040 和 3GPP TS 23.038 规范。

最初的设计目的是提供通过 GSM 调制解调器发送和接收短信的功能，但该库在任何需要编码和解码 SMS TPDU 或其字段的场景中都可以使用。

## 致谢

本库基于 [warthog618/sms](https://github.com/warthog618/sms) 进行修改。由于原贡献者已不再维护该包，而项目需要根据实际需求进行一些细节调整，因此复制并修改了原始代码。感谢原作者做出的杰出贡献。

## 功能特性

支持以下功能：

- 从 UTF-8 字符串创建 SMS TPDU，包括表情符号 😁
- 将长消息分割为多个连续的 SMS TPDU
- 编码时自动选择字符集和语言
- 将 SMS TPDU 解码为 UTF-8 字符串
- 将连续的 SMS TPDU 重新组合为完整的长消息
- 支持所有 GSM 字符集
- 支持 PDU 模式的 SMS TPDU 编码和解码，用于与 GSM 调制解调器交换数据

## 使用方法

```bash
go get github.com/rehiy/modem
```

### 编码

创建包含消息的 TPDU 称为编码。

可以使用 `sms.Encode` 编码单条消息：

```go
import "github.com/rehiy/modem/sms"

tpdus, _ := sms.Encode([]byte("hello world"))
for _, p := range tpdus {
    b, _ := p.MarshalBinary()
    // 发送二进制 TPDU...
}
```

发送多条消息需要维护多个计数字段并在 TPDU 中编码它们。这可以通过 `sms.Encoder` 完成：

```go
e := sms.NewEncoder()
for {
    msg := <- msgChan
    tpdus, _ := e.Encode(msg)
    for _, p := range tpdus {
        b, _ := p.MarshalBinary()
        // 发送二进制 TPDU...
    }
}
```

### 反序列化

将接收到的 TPDU 重新组合为完整消息是一个多步骤过程。第一步是使用 `sms.Unmarshal` 将二进制 SMS TPDU 反序列化为 TPDU 对象：

```go
pdu, _ := sms.Unmarshal(bintpdu)
```

### 解码

单个分段 TPDU 可以使用 `sms.Decode` 解码：

```go
msg, _ := sms.Decode([]*tpdu.TPDUs{pdu})
```

对于连续消息，使用 `sms.Decode` 将包含消息的一组 TPDU 重新组合为完整消息：

```go
msg, _ := sms.Decode(tpdus)
```

### 收集

连续消息的分段必须在解码前收集。Collector 收集接收到的分段，并在接收到最后一个分段时返回完整的集合。

```go
c := sms.NewCollector()
for {
    bintpdu := <- pduChan
    pdu, _ := sms.Unmarshal(bintpdu)
    tpdus, _ := c.Collect(pdu)
    if len(tpdus) > 0 {
        msg, _ := sms.Decode(tpdus)
        // 处理消息...
    }
}
```

### 选项

核心 API 旨在满足最常见的用例，即针对移动台的操作。例如，默认情况下 `sms.Encode` 创建 SMS-SUBMIT TPDU 并仅使用默认字符集。默认情况下，`sms.Decode` 使用所有字符集。默认情况下，`sms.Unmarshal` 假设 TPDU 是移动终端。

可以使用可选参数为核心 API 函数的行为进行其他用例的调整。

例如，为 SMS-SUBMIT 消息指定目标号码：

```go
tpdus, _ := sms.Encode("hello",sms.To("12345"))
```

或者在必要时使用特定字符集编码消息：

```go
tpdus, _ := sms.Encode("hello ٻ",sms.WithCharset(charset.Urdu))
```

或者指定 SMS-DELIVER 消息的编码：

```go
tpdus, _ := sms.Encode("hello",sms.AsDeliver,sms.From("12345"))
```

或者从移动站反序列化 TPDU：

```go
pdu, _ := sms.Unmarshal(bintpdu,sms.AsMO)
```

完整提供的选项列表：

选项 | 类别 | 描述
---|---|---
*WithReassemblyTimeout(duration,handler)*|Collect|限制等待完整重新组装的 TPDU 的时间
*WithTemplate(tpdu)*|Encode|使用提供的 TPDU 作为编码 TPDU 的模板
*WithTemplateOption(tpdu.Option)*|Encode|在编码期间将提供的选项应用于模板 TPDU
*To(number)*|Encode|将编码 TPDU 的 DA 设置为提供的号码
*From(number)*|Encode|将编码 TPDU 的 OA 设置为提供的号码
*WithAllCharsets*|Decode,Encode|使所有 GSM7 字符集可用
*WithDefaultCharset*|Decode,Encode|仅使默认字符集可用
*WithCharset(nli...)*|Decode,Encode|使指定的字符集可用
*WithLockingCharset(nli...)*|Decode,Encode|使指定的字符集可作为锁定字符集使用
*WithShiftCharset(nli...)*|Decode,Encode|使指定的字符集可作为移位字符集使用
*AsSubmit*|Encode|将 TPDU 编码为 SMS-SUBMIT（默认）
*AsDeliver*|Encode|将 TPDU 编码为 SMS-DELIVER
*As8Bit*|Encode|强制将用户数据编码为 8 位
*AsUCS2*|Encode|强制将用户数据编码为 UCS-2
*AsMO*|Unmarshal|将 TPDU 视为从移动台发起
*AsMT*|Unmarshal|将 TPDU 视为在移动台终止（默认）
