package observe

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"main.go/global"
	"main.go/model/observe"
)

// MapStatus 状态映射 (firing → 告警中, resolved → 已恢复)
func MapStatus(status string) string {
	switch status {
	case "firing":
		return "告警中"
	case "resolved":
		return "已恢复"
	default:
		return status
	}
}

// MapObjectKind 对象类型映射(K8s常见资源)
func MapObjectKind(kind string) string {
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

// ParseI18nField 解析国际化字段，返回中文值
// 支持两种格式:
// - JSON格式: {"zh":"CPU使用率高","en":"High CPU usage"}
// - 纯文本格式: 应用连续3分钟调度失败
func ParseI18nField(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}

	// 检查是否是 JSON 格式（以 { 开头）
	trimmed := strings.TrimSpace(jsonStr)
	if !strings.HasPrefix(trimmed, "{") {
		// 不是 JSON，直接返回原值
		return jsonStr
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

// GetDisplayName 获取告警显示名称（带回退逻辑）
// 优先级: ParseI18nField(DisplayName) → AlertIndicatorAlias → AlertIndicator
// 同时去除与 objectKindZh 重复的前缀
func GetDisplayName(labels observe.AlertLabels, objectKindZh string) string {
	displayNameZh := ParseI18nField(labels.DisplayName)

	// 如果 displayName 为空，使用 alert_indicator_alias 作为备选
	if displayNameZh == "" {
		displayNameZh = labels.AlertIndicatorAlias
	}
	// 如果 alert_indicator_alias 也为空，使用 alert_indicator
	if displayNameZh == "" {
		displayNameZh = labels.AlertIndicator
	}

	// 去除 displayNameZh 中重复的 objectKindZh 前缀
	if strings.HasPrefix(displayNameZh, objectKindZh) {
		displayNameZh = strings.TrimPrefix(displayNameZh, objectKindZh)
	}

	return displayNameZh
}

// BuildAlertObject 构建告警对象
// 格式: objectKindZh + AlertInvolvedObjectName
func BuildAlertObject(labels observe.AlertLabels) string {
	objectKindZh := MapObjectKind(labels.AlertInvolvedObjectKind)
	return objectKindZh + labels.AlertInvolvedObjectName
}

// BuildAlertDesc 构建告警描述
// 格式: alertObject + displayNameZh + AlertIndicatorComparison + AlertIndicatorThreshold
func BuildAlertDesc(labels observe.AlertLabels) string {
	objectKindZh := MapObjectKind(labels.AlertInvolvedObjectKind)
	alertObject := BuildAlertObject(labels)
	displayNameZh := GetDisplayName(labels, objectKindZh)
	return alertObject + " " + displayNameZh + " " + labels.AlertIndicatorComparison + " " + labels.AlertIndicatorThreshold
}

// BuildEmailSubject 构建邮件主题
func BuildEmailSubject(status string, labels observe.AlertLabels) string {
	statusZh := MapStatus(status)
	alertDesc := BuildAlertDesc(labels)

	// 邮件主题 = 【状态】PAAS 平台告警：+ 告警描述
	return fmt.Sprintf("【%s】PAAS 平台告警：%s", statusZh, alertDesc)
}
