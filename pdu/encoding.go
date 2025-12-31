package pdu

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf16"
)

// GSM 7-bit 默认字符集（3GPP TS 23.038）
var gsm7bitChars = "@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞ\x1bÆæßÉ !\"#¤%&'()*+,-./0123456789:;<=>?¡ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÑÜ§¿abcdefghijklmnopqrstuvwxyzäöñüà"

// GSM 7-bit 扩展字符集（需要转义字符 0x1B）
var gsm7bitExtChars = map[rune]byte{
	'|': 0x40, '^': 0x14, '€': 0x65, '{': 0x28, '}': 0x29,
	'[': 0x3C, ']': 0x3E, '~': 0x3D, '\\': 0x2F,
}

// gsm7bitExtCharsReverse 扩展字符反向映射表，用于解码时 O(1) 查找
var gsm7bitExtCharsReverse map[byte]rune

func init() {
	gsm7bitExtCharsReverse = make(map[byte]rune, len(gsm7bitExtChars))
	for r, b := range gsm7bitExtChars {
		gsm7bitExtCharsReverse[b] = r
	}
}

// Encode7Bit 将文本编码为 GSM 7-bit 格式
// 扩展字符会被编码为两个字节：0x1B + 扩展码
func Encode7Bit(text string) ([]byte, error) {
	septets := make([]byte, 0, len(text))
	gsm7bitRunes := []rune(gsm7bitChars)

	for _, r := range text {
		if extCode, ok := gsm7bitExtChars[r]; ok {
			septets = append(septets, 0x1B, extCode)
			continue
		}
		// 在rune数组中查找字符位置
		index := -1
		for i, c := range gsm7bitRunes {
			if c == r {
				index = i
				break
			}
		}
		if index == -1 {
			return nil, fmt.Errorf("character '%c' not supported in GSM 7-bit", r)
		}
		septets = append(septets, byte(index))
	}
	return pack7Bit(septets), nil
}

// Decode7Bit 解码 GSM 7-bit 数据
// length 参数指定要解码的字符数（septets）
func Decode7Bit(data []byte, length int) string {
	septets := unpack7Bit(data, length)
	gsm7bitRunes := []rune(gsm7bitChars)
	var result strings.Builder
	result.Grow(length)
	escape := false

	for _, septet := range septets {
		if escape {
			if r, ok := gsm7bitExtCharsReverse[septet]; ok {
				result.WriteRune(r)
			}
			escape = false
		} else if septet == 0x1B {
			escape = true
		} else if int(septet) < len(gsm7bitRunes) {
			result.WriteRune(gsm7bitRunes[septet])
		}
	}
	return result.String()
}

// pack7Bit 将 7-bit septets 打包为 8-bit 字节
// GSM 7-bit 编码将每 8 个 7-bit 字符打包为 7 个字节
func pack7Bit(septets []byte) []byte {
	if len(septets) == 0 {
		return []byte{}
	}

	packed := make([]byte, 0, (len(septets)*7+7)/8)
	bits := uint(0)
	buffer := uint32(0)

	for _, septet := range septets {
		buffer |= uint32(septet) << bits
		bits += 7

		for bits >= 8 {
			packed = append(packed, byte(buffer&0xFF))
			buffer >>= 8
			bits -= 8
		}
	}

	if bits > 0 {
		packed = append(packed, byte(buffer&0xFF))
	}

	return packed
}

// unpack7Bit 将 8-bit 字节解包为 7-bit septets
// 这是 pack7Bit 的逆操作
func unpack7Bit(data []byte, length int) []byte {
	if len(data) == 0 || length == 0 {
		return []byte{}
	}

	septets := make([]byte, 0, length)
	bits := uint(0)
	buffer := uint32(0)

	for _, b := range data {
		buffer |= uint32(b) << bits
		bits += 8

		for bits >= 7 && len(septets) < length {
			septets = append(septets, byte(buffer&0x7F))
			buffer >>= 7
			bits -= 7
		}

		if len(septets) >= length {
			break
		}
	}

	return septets
}

// EncodeUCS2 将文本编码为 UCS2（UTF-16 Big Endian）
func EncodeUCS2(text string) []byte {
	runes := []rune(text)
	utf16Codes := utf16.Encode(runes)
	result := make([]byte, len(utf16Codes)*2)

	for i, code := range utf16Codes {
		result[i*2] = byte(code >> 8)
		result[i*2+1] = byte(code)
	}

	return result
}

// DecodeUCS2 解码 UCS2（UTF-16 Big Endian）数据
func DecodeUCS2(data []byte) string {
	if len(data)%2 != 0 {
		return ""
	}

	utf16Codes := make([]uint16, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		utf16Codes[i/2] = uint16(data[i])<<8 | uint16(data[i+1])
	}

	return string(utf16.Decode(utf16Codes))
}

// SwapNibbles 交换字符串中每对字符的位置
// 例如："1234" -> "2143"
// 用于电话号码的 BCD 编码
func SwapNibbles(s string) string {
	if len(s) == 0 {
		return s
	}

	bytes := []byte(s)
	for i := 0; i < len(bytes)-1; i += 2 {
		bytes[i], bytes[i+1] = bytes[i+1], bytes[i]
	}
	return string(bytes)
}

// EncodePhoneNumber 编码电话号码为 BCD 格式
// 返回地址类型和交换后的十六进制字符串
func EncodePhoneNumber(number string) (AddressType, string) {
	var cleaned strings.Builder
	cleaned.Grow(len(number))
	international := false

	for _, r := range number {
		if r == '+' {
			international = true
		} else if r >= '0' && r <= '9' {
			cleaned.WriteRune(r)
		}
	}

	result := cleaned.String()
	if len(result)%2 != 0 {
		result += "F"
	}

	addrType := AddressTypeUnknown
	if international {
		addrType = AddressTypeInternational
	}

	return addrType, SwapNibbles(result)
}

// DecodePhoneNumber 解码 BCD 格式的电话号码
func DecodePhoneNumber(data string, addrType AddressType) string {
	// 字母数字地址：直接 7-bit 解码（不进行 nibble 交换）
	if addrType == AddressTypeAlphanumeric {
		bytes, err := HexToBytes(data)
		if err != nil {
			return data
		}
		// 计算字符数：addrLen 通常包含在调用处
		// 这里简化处理，使用 unpack7Bit 的默认长度
		return Decode7Bit(bytes, (len(bytes)*8)/7)
	}

	// BCD 编码的电话号码，需要交换半字节
	swapped := SwapNibbles(data)
	// 移除填充的 F（可能在末尾）
	swapped = strings.TrimSuffix(swapped, "F")

	if addrType == AddressTypeInternational {
		return "+" + swapped
	}
	return swapped
}

func HexToBytes(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("hex string length must be even")
	}
	return hex.DecodeString(hexStr)
}

func BytesToHex(bytes []byte) string {
	return strings.ToUpper(hex.EncodeToString(bytes))
}
