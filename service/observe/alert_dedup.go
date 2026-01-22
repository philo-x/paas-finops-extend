package observe

import (
	"crypto/md5"
	"fmt"
	"time"

	"main.go/global"
	"main.go/model/observe"
)

type AlertDedupService struct {
}

// GenerateFingerprint 生成告警指纹
// 使用 alert_cluster + alert_involved_object_kind + alert_involved_object_name + alert_resource + alertname 组合生成MD5
func (s *AlertDedupService) GenerateFingerprint(labels observe.AlertLabels) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s",
		labels.AlertCluster,
		labels.AlertInvolvedObjectKind,
		labels.AlertInvolvedObjectName,
		labels.AlertResource,
		labels.Alertname,
	)
	hash := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", hash)
}

// FindActiveAlertByFingerprint 根据指纹查找活跃告警(status=firing, is_deleted=0)
func (s *AlertDedupService) FindActiveAlertByFingerprint(fingerprint string) (*observe.PrometheusAlert, error) {
	var alert observe.PrometheusAlert
	err := global.GVA_DB.Where("fingerprint = ? AND status = ? AND is_deleted = 0", fingerprint, "firing").First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// ShouldSendNotification 判断是否应该发送通知
// resolved状态始终发送，firing状态根据每日限制判断
func (s *AlertDedupService) ShouldSendNotification(alert *observe.PrometheusAlert, isResolved bool) bool {
	// resolved状态始终发送通知
	if isResolved {
		return true
	}

	// 获取每日通知限制配置
	dailyLimit := global.GVA_CONFIG.MQ.DailyNotifyLimit
	// 如果限制为0，表示不限制
	if dailyLimit <= 0 {
		return true
	}

	today := time.Now().Truncate(24 * time.Hour)

	// 检查是否同一天
	if alert.LastNotifyDate != nil {
		lastDate := alert.LastNotifyDate.Truncate(24 * time.Hour)
		if lastDate.Equal(today) {
			// 同一天，检查是否超过限制
			return alert.DailyNotifyCount < dailyLimit
		}
	}

	// 不同一天或者第一次，可以发送
	return true
}

// UpdateNotifyCount 更新通知计数
func (s *AlertDedupService) UpdateNotifyCount(alert *observe.PrometheusAlert) {
	today := time.Now().Truncate(24 * time.Hour)

	// 检查是否同一天
	if alert.LastNotifyDate != nil {
		lastDate := alert.LastNotifyDate.Truncate(24 * time.Hour)
		if lastDate.Equal(today) {
			// 同一天，递增计数
			alert.DailyNotifyCount++
		} else {
			// 不同一天，重置计数
			alert.DailyNotifyCount = 1
			alert.LastNotifyDate = &today
		}
	} else {
		// 第一次通知
		alert.DailyNotifyCount = 1
		alert.LastNotifyDate = &today
	}
}

// ResetDailyNotifyCount 重置每日通知计数(跨天时调用)
func (s *AlertDedupService) ResetDailyNotifyCount(alert *observe.PrometheusAlert) {
	today := time.Now().Truncate(24 * time.Hour)
	alert.DailyNotifyCount = 1
	alert.LastNotifyDate = &today
}
