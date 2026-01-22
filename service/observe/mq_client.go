package observe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	statusZh := s.mapStatus(alert.Status)
	levelZh := s.mapSeverity(alert.Labels.Severity)
	objectKindZh := s.mapObjectKind(alert.Labels.AlertInvolvedObjectKind)
	displayNameZh := s.parseI18nField(alert.Labels.DisplayName)

	// 如果 displayName 为空，使用 alert_indicator_alias 作为备选
	if displayNameZh == "" {
		displayNameZh = alert.Labels.AlertIndicatorAlias
	}
	// 如果 alert_indicator_alias 也为空，使用 alert_indicator
	if displayNameZh == "" {
		displayNameZh = alert.Labels.AlertIndicator
	}

	// 去除 displayNameZh 中重复的 objectKindZh 前缀
	if strings.HasPrefix(displayNameZh, objectKindZh) {
		displayNameZh = strings.TrimPrefix(displayNameZh, objectKindZh)
	}

	// 构建告警对象
	alertObject := objectKindZh + alert.Labels.AlertInvolvedObjectName

	// 告警描述 = 告警对象 + displayName(zh优先) + alert_indicator_comparison + alert_indicator_threshold
	alertDesc := alertObject + displayNameZh + alert.Labels.AlertIndicatorComparison + alert.Labels.AlertIndicatorThreshold

	// 邮件主题 = 【状态】PAAS 平台告警：+ 告警描述
	emailSubject := fmt.Sprintf("【%s】PAAS 平台告警：%s", statusZh, alertDesc)

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

// mapStatus 状态映射
func (s *MQClientService) mapStatus(status string) string {
	switch status {
	case "firing":
		return "告警中"
	case "resolved":
		return "已恢复"
	default:
		return status
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

// mapObjectKind 对象类型映射(K8s常见资源)
func (s *MQClientService) mapObjectKind(kind string) string {
	switch kind {
	// 集群
	case "Cluster":
		return "集群"
	// 工作负载资源
	case "Node":
		return "节点"
	case "Pod":
		return "Pod"
	case "Deployment":
		return "部署"
	case "StatefulSet":
		return "有状态副本集"
	case "DaemonSet":
		return "守护进程集"
	case "ReplicaSet":
		return "副本集"
	case "Job":
		return "任务"
	case "CronJob":
		return "定时任务"
	// 服务发现与负载均衡
	case "Service":
		return "服务"
	case "Ingress":
		return "入口"
	case "Endpoints":
		return "端点"
	// 配置与存储
	case "ConfigMap":
		return "配置字典"
	case "Secret":
		return "密钥"
	case "PersistentVolume":
		return "持久卷"
	case "PersistentVolumeClaim":
		return "持久卷声明"
	case "StorageClass":
		return "存储类"
	// 命名空间与资源配额
	case "Namespace":
		return "命名空间"
	case "ResourceQuota":
		return "资源配额"
	case "LimitRange":
		return "限制范围"
	// 访问控制
	case "ServiceAccount":
		return "服务账号"
	case "Role":
		return "角色"
	case "ClusterRole":
		return "集群角色"
	case "RoleBinding":
		return "角色绑定"
	case "ClusterRoleBinding":
		return "集群角色绑定"
	// 网络策略
	case "NetworkPolicy":
		return "网络策略"
	// 自定义资源
	case "HorizontalPodAutoscaler":
		return "水平自动伸缩"
	case "VerticalPodAutoscaler":
		return "垂直自动伸缩"
	case "PodDisruptionBudget":
		return "Pod中断预算"
	default:
		return kind
	}
}

// parseI18nField 解析国际化字段，返回中文值
func (s *MQClientService) parseI18nField(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}

	var i18n struct {
		Zh string `json:"zh"`
		En string `json:"en"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &i18n); err != nil {
		// 解析失败时记录日志，便于排查问题
		global.GVA_LOG.Warn("解析 i18n 字段失败",
			zap.String("input", jsonStr),
			zap.Error(err))
		return jsonStr
	}

	if i18n.Zh != "" {
		return i18n.Zh
	}
	return i18n.En
}
