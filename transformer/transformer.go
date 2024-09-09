package transformer

import (
	"bytes"
	"fmt"

	"github.com/k8stech/alertmanager-wechatrobot-webhook/model"
)

// 新增一个函数来获取告警颜色
func getAlertColor(severity string) string {
	switch severity {
	case "critical":
		return "warning"
	case "warning":
		return "warning"
	case "resolved":
		return "info"
	default:
		return "comment"
	}
}

// TransformToMarkdown transform alertmanager notification to wechat markdow message
func TransformToMarkdown(notification model.Notification, grafanaURL string, alertDomain string) (markdown *model.WeChatMarkdown, robotURL string, err error) {

	status := notification.Status

	annotations := notification.CommonAnnotations
	robotURL = annotations["wechatRobot"]

	var buffer bytes.Buffer

	//buffer.WriteString(fmt.Sprintf("### 当前状态:%s \n", status))
	// buffer.WriteString(fmt.Sprintf("#### 告警项:\n"))

	for _, alert := range notification.Alerts {
		labels := alert.Labels

		// 动态获取 var-NameSpace 和 var-Container
		namespace := labels["namespace"]
		pod := labels["pod"]
		// 获取告警等级
		severity := labels["severity"]
		// 获取对应的颜色
		alertColor := getAlertColor(severity)

		buffer.WriteString(fmt.Sprintf("### 当前状态: <font color='%s'> %s </font>\n", alertColor, status))

		fmt.Printf("NOVA namespace:%s, pod: %s, alertDomain: %s\n", namespace, pod, alertDomain)
		buffer.WriteString(fmt.Sprintf("\n # 告警: <font color='%s'> %s </font>\n", alertColor, annotations["summary"]))
		// datacenter 为 victoriametrics 获取 prometheus时区分环境的 label
		//buffer.WriteString(fmt.Sprintf("\n>【环境】 %s\n", labels["datacenter"]))
		buffer.WriteString(fmt.Sprintf("\n>【级别】 %s\n", severity))
		buffer.WriteString(fmt.Sprintf("\n>【类型】 %s\n", labels["alertname"]))
		buffer.WriteString(fmt.Sprintf("\n>【主机】 %s\n", labels["instance"]))
		buffer.WriteString(fmt.Sprintf("\n>【内容】 %s\n", alert.Annotations["description"]))
		buffer.WriteString(fmt.Sprintf("\n>【当前状态】%s \n", status))
		buffer.WriteString(fmt.Sprintf("\n>【触发时间】 %s\n", alert.StartsAt.Format("2006-01-02 15:04:05")))
		//buffer.WriteString(fmt.Sprintf("\n [跳转Grafana看板](https://grafana.vnnox.com/d/PwMJtdvnr/k8s-chu-neng-cnanduat?orgId=1&var-origin_prometheus=&var-Node=All&var-NameSpace=%s&var-Container=%s&var-Pod=All)", namespace, container))
		buffer.WriteString(fmt.Sprintf("\n [跳转Grafana看板](https://%s?orgId=1&var-origin_prometheus=&var-Node=All&var-NameSpace=%s&var-Pod=%s&var-Pod=All)", grafanaURL, namespace, pod))
		buffer.WriteString(fmt.Sprintf("\n [告警规则详情](http://%s/alerts?search=)", alertDomain))
		buffer.WriteString(fmt.Sprintf("\n [日志详情](https://aws-au-loki-grafana.vnnox.com/d/o6-BGgnnk/loki-kubernetes-logs?orgId=1&from=now-1h&to=now&var-query=&var-namespace=au&var-stream=All&var-container=vnnox-middle-oauth)"))
		buffer.WriteString(fmt.Sprintf("\n"))
		buffer.WriteString(fmt.Sprintf("---------------------------------------------------------------------------\n"))

		//buffer.WriteString("\n [告警规则详情](http://%s/alerts?search=)", alertDomain)
		//buffer.WriteString(fmt.Sprintf("\n>告警级别: %s\n", labels["severity"]))
		//buffer.WriteString(fmt.Sprintf("\n>告警类型: %s\n", labels["alertname"]))
		//buffer.WriteString(fmt.Sprintf("\n>故障主机: %s\n", labels["instance"]))
		//
		//annotations := alert.Annotations
		//buffer.WriteString(fmt.Sprintf("\n>告警主题: %s\n", annotations["summary"]))
		//buffer.WriteString(fmt.Sprintf("\n>告警详情: %s\n", annotations["description"]))
		//buffer.WriteString(fmt.Sprintf("\n> 触发时间: %s\n", alert.StartsAt.Format("2006-01-02 15:04:05")))
	}

	markdown = &model.WeChatMarkdown{
		MsgType: "markdown",
		Markdown: &model.Markdown{
			Content: buffer.String(),
		},
	}

	return
}
