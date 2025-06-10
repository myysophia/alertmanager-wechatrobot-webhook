package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var (
	kongAPIEndpoint string
	defaultTimeout  int
	regionEndpoints = map[string]string{
		"us": "https://prometheus-kong-us.vnnox.com/request-limit/add",
		"in": "https://prometheus-kong-in.vnnox.com/request-limit/add",
		"eu": "https://prometheus-kong-eu.vnnox.com/request-limit/add",
		"cn": "https://prometheus-kong-cn.vnnox.com/request-limit/add",
		"au": "https://prometheus-kong-au.vnnox.com/request-limit/add",
	}
)

func init() {
	// 从环境变量读取配置，如果未设置则使用默认值
	kongAPIEndpoint = getEnvOrDefault("KONG_API_ENDPOINT", "https://192.168.2.79:8003/request-limit/add")

	// 从环境变量读取timeout配置，如果未设置则使用默认值600
	timeoutStr := getEnvOrDefault("KONG_TIMEOUT", "600")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = 600 // 如果转换失败，使用默认值
	}
	defaultTimeout = timeout
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
func HandleSMSPrefixBlock(prefix string, region string) error {
	// 根据region获取对应的Kong API endpoint
	endpoint, exists := regionEndpoints[region]
	if !exists {
		// 如果region不存在，使用默认endpoint
		fmt.Printf("Warning: Unknown region %s, using default endpoint\n", region)
		endpoint = kongAPIEndpoint
	}

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
				Timeout: defaultTimeout,
			},
		},
	}

	// 转换为JSON
	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("marshal kong rule failed: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
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
	fmt.Printf("The API response body is : %s\n", resp.Body)
	fmt.Printf("Kong API request successful with status: %s\n", resp.Status)
	return nil
}

// SendBlockSuccessNotification 发送封禁成功的通知
func SendBlockSuccessNotification(prefix string, robotURL string, defaultRobot string) error {
	content := fmt.Sprintf("### 短信前缀封禁成功\n"+
		"**前缀**: %s\n"+
		"**操作**: 已成功添加到Kong封禁规则\n"+
		"**封禁时长**: %d秒", prefix, defaultTimeout)

	return sendToWeChat(content, robotURL, defaultRobot)
}
