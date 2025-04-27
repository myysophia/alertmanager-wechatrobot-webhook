package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var (
	kongAPIEndpoint string
)

func init() {
	// 从环境变量读取配置，如果未设置则使用默认值
	kongAPIEndpoint = getEnvOrDefault("KONG_API_ENDPOINT", "https://192.168.2.79:8003/request-limit/add")
}

// getEnvOrDefault 从环境变量获取值，如果未设置则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type KongRule struct {
	Action string `json:"action"`
	Rules  []struct {
		Type string `json:"type"`
		Rule struct {
			Prefix string `json:"prefix"`
		} `json:"rule"`
		Timeout int `json:"timeout"`
	} `json:"rules"`
}

// HandleSMSPrefixBlock 处理短信前缀封禁逻辑
func HandleSMSPrefixBlock(prefix string) error {
	// 构建请求体
	rule := KongRule{
		Action: "add",
		Rules: []struct {
			Type string `json:"type"`
			Rule struct {
				Prefix string `json:"prefix"`
			} `json:"rule"`
			Timeout int `json:"timeout"`
		}{
			{
				Type: "sms",
				Rule: struct {
					Prefix string `json:"prefix"`
				}{
					Prefix: prefix,
				},
				Timeout: 600,
			},
		},
	}

	// 转换为JSON
	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("marshal kong rule failed: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", kongAPIEndpoint, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kong API request failed with status: %s", resp.Status)
	}
	fmt.Printf("Kong API request successful with status: %s\n", resp.Status)
	return nil
}

// SendBlockSuccessNotification 发送封禁成功的通知
func SendBlockSuccessNotification(prefix string, robotURL string, defaultRobot string) error {
	content := fmt.Sprintf("### 短信前缀封禁成功\n"+
		"**前缀**: %s\n"+
		"**操作**: 已成功添加到Kong封禁规则\n"+
		"**封禁时长**: 600秒", prefix)

	return sendToWeChat(content, robotURL, defaultRobot)
}
