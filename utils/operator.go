package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Operator 表示运营商信息
type Operator struct {
	MCC     string `json:"mcc,omitempty"`     // 移动国家代码
	MNC     string `json:"mnc,omitempty"`     // 移动网络代码
	Country string `json:"country,omitempty"` // 国家名称
	Code    string `json:"code,omitempty"`    // 国家代码 (ISO 3166-1 alpha-2)
	Name    string `json:"name,omitempty"`    // 运营商名称
	Brand   string `json:"brand,omitempty"`   // 品牌名称
	Network string `json:"network,omitempty"` // 网络类型 (GSM, LTE, etc.)
	Status  string `json:"status,omitempty"`  // 状态 (active, inactive)
	Note    string `json:"note,omitempty"`    // 备注
}

// QueryPLMN 通过 PLMN、国家代码或模糊搜索查询运营商信息
// 参数 arg 可以是 PLMN (如 "46001")、ISO 国家代码 (如 "CN") 或模糊搜索词 (如 "China Mobile")
// 返回 Operator 切片和错误信息。对于精确查询返回单个元素的切片，模糊搜索返回多个结果。
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

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("API 返回空响应")
	}

	var op *Operator
	err = json.Unmarshal(body, op)
	return op, err
}
