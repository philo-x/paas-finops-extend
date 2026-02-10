package observe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"main.go/global"
	"main.go/model/observe"
)

// mqHTTPClient 全局 HTTP 客户端，支持连接池复用
var mqHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

// initMQClient 根据配置初始化 MQ HTTP 客户端
func initMQClient() {
	if global.GVA_CONFIG.MQ.Timeout > 0 {
		mqHTTPClient.Timeout = time.Duration(global.GVA_CONFIG.MQ.Timeout) * time.Second
	}
}

type MQClientService struct{}

// SendAlertNotification 发送告警通知到MQ
func (s *MQClientService) SendAlertNotification(alert observe.PrometheusAlert) error {
	// 确保 HTTP 客户端已初始化
	initMQClient()

	mqMsg := s.buildMQMessage(alert)

	jsonData, err := json.Marshal(mqMsg)
	if err != nil {
		return fmt.Errorf("序列化MQ消息失败: %w", err)
	}

	global.GVA_LOG.Info("MQ消息序列化完成", zap.String("jsonData", string(jsonData)))

	resp, err := mqHTTPClient.Post(global.GVA_CONFIG.MQ.Url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送MQ请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MQ返回非200状态码: %d", resp.StatusCode)
	}
	return nil
}

// buildMQMessage 构建MQ消息体
func (s *MQClientService) buildMQMessage(alert observe.PrometheusAlert) observe.MQMessageRequest {
	statusZh := MapStatus(alert.Status)
	levelZh := s.mapSeverity(alert.Labels.Severity)

	// 使用共享函数构建各字段
	emailSubject := BuildEmailSubject(alert.Status, alert.Labels)
	alertObject := BuildAlertObject(alert.Labels)
	alertDesc := BuildAlertDesc(alert.Labels)

	return observe.MQMessageRequest{
		Topic: global.GVA_CONFIG.MQ.Topic,
		Tag:   global.GVA_CONFIG.MQ.Tag,
		Data: observe.MQAlertData{
			Title:    emailSubject,
			Receiver: global.GVA_CONFIG.MQ.Receiver,
			AlertDetail: observe.MQAlertDetail{
				Status:       statusZh,
				Severity:     levelZh,
				Cluster:      alert.Labels.AlertCluster,
				Object:       alertObject,
				Indicator:    alert.Labels.AlertResource,
				Summary:      alertDesc,
				TriggerValue: alert.Annotations.AlertCurrentValue,
				AlertTime:    alert.StartsAt.Format("2006-01-02 15:04:05"),
				Remark:       "",
			},
		},
	}
}

// mapSeverity 等级映射
func (s *MQClientService) mapSeverity(severity string) string {
	switch severity {
	case "Critical":
		return "紧急"
	case "High":
		return "严重"
	case "Warning":
		return "警告"
	case "Low":
		return "轻微"
	default:
		return "一般"
	}
}
