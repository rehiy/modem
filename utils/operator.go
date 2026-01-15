package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Operator 表示运营商信息
type Operator struct {
	MCC      int    `json:"mcc,omitempty"`      // 移动国家代码
	MNC      int    `json:"mnc,omitempty"`      // 移动网络代码
	PLMN     int    `json:"plmn,omitempty"`     // PLMN代码
	Region   string `json:"region,omitempty"`   // 地区
	Country  string `json:"country,omitempty"`  // 国家名称
	ISO      string `json:"iso,omitempty"`      // ISO国家代码
	Operator string `json:"operator,omitempty"` // 运营商名称
	Brand    string `json:"brand,omitempty"`    // 品牌名称
	TADIG    string `json:"tadig,omitempty"`    // TADIG代码
	Bands    string `json:"bands,omitempty"`    // 频段信息
	Network  string `json:"network,omitempty"`  // 网络类型 (GSM, LTE, etc.)
	Status   string `json:"status,omitempty"`   // 状态 (active, inactive)
	Note     string `json:"note,omitempty"`     // 备注
}

// QueryPLMN 通过 PLMN、国家代码或模糊搜索查询运营商信息
// 参数 arg 可以是 PLMN (如 "46001")、ISO 国家代码 (如 "CN") 或模糊搜索词 (如 "China Mobile")
// 返回 Operator 指针和错误信息。API总是返回单个对象。
func QueryPLMN(arg string) (*Operator, error) {
	url := fmt.Sprintf("https://api.rehi.org/plmn/%s", arg)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) == 0 {
		return nil, fmt.Errorf("API 返回空响应")
	}

	if bodyStr[0] != '{' {
		previewLen := 20
		if len(bodyStr) < previewLen {
			previewLen = len(bodyStr)
		}
		return nil, fmt.Errorf("无效的JSON响应: %s", bodyStr[:previewLen])
	}

	var op Operator
	if err := json.Unmarshal([]byte(bodyStr), &op); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &op, nil
}
