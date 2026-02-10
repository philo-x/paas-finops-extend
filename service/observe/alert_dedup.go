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
// 基于告警描述生成MD5（不包含状态），格式: 告警对象+displayName+comparison+threshold
func (s *AlertDedupService) GenerateFingerprint(labels observe.AlertLabels) string {
	alertDesc := BuildAlertDesc(labels)
	hash := md5.Sum([]byte(alertDesc))
	return fmt.Sprintf("%x", hash)
}

// FindAlertByFingerprint 根据指纹查找告警(所有状态统一去重, is_deleted=0)
func (s *AlertDedupService) FindAlertByFingerprint(fingerprint string) (*observe.PrometheusAlert, error) {
	var alert observe.PrometheusAlert
	err := global.GVA_DB.Where("fingerprint = ? AND is_deleted = 0", fingerprint).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// ShouldSendNotification 判断是否应该发送通知
// 如果NotifyPending=true则始终返回true(上次发送失败,需要重试)
// firing和resolved状态都根据每日限制判断
func (s *AlertDedupService) ShouldSendNotification(alert *observe.PrometheusAlert) bool {
	// 如果有待发送的通知(上次发送失败)，始终尝试发送
	if alert.NotifyPending {
		return true
	}

	// 获取每日通知限制配置
	dailyLimit := global.GVA_CONFIG.MQ.DailyNotifyLimit
	// 如果限制为0，表示不限制
	if dailyLimit <= 0 {
		return true
	}

	now := time.Now()

	// 检查是否同一天(使用本地时区的年月日比较，避免Truncate的UTC时区问题)
	if alert.LastNotifyDate != nil {
		if isSameDay(now, *alert.LastNotifyDate) {
			// 同一天，检查是否超过限制
			return alert.DailyNotifyCount < dailyLimit
		}
	}

	// 不同一天或者第一次，可以发送
	return true
}

// isSameDay 比较两个时间是否是同一天(基于本地时区)
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Local().Date()
	y2, m2, d2 := t2.Local().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// UpdateNotifyCount 设置待发送通知标记
// 注意:不再直接增加DailyNotifyCount,而是设置NotifyPending=true
// DailyNotifyCount的增加移到ConfirmNotifySent中(发送成功后才增加)
func (s *AlertDedupService) UpdateNotifyCount(alert *observe.PrometheusAlert) {
	// 设置待发送通知标记
	alert.NotifyPending = true
}

// ConfirmNotifySent 确认通知发送成功
// 发送成功后调用此方法: 清除NotifyPending标记并原子增加DailyNotifyCount
func (s *AlertDedupService) ConfirmNotifySent(alertId int) error {
	now := time.Now()

	// 先查询当前告警以获取LastNotifyDate
	var alert observe.PrometheusAlert
	if err := global.GVA_DB.Where("alert_id = ?", alertId).First(&alert).Error; err != nil {
		return err
	}

	// 判断是否需要重置计数(跨天)
	updates := map[string]interface{}{
		"notify_pending": false,
	}

	if alert.LastNotifyDate != nil && isSameDay(now, *alert.LastNotifyDate) {
		// 同一天，使用原子增加
		return global.GVA_DB.Model(&observe.PrometheusAlert{}).
			Where("alert_id = ?", alertId).
			Updates(updates).
			UpdateColumn("daily_notify_count", global.GVA_DB.Raw("daily_notify_count + 1")).Error
	} else {
		// 不同一天或第一次，重置计数为1
		updates["daily_notify_count"] = 1
		updates["last_notify_date"] = now
		return global.GVA_DB.Model(&observe.PrometheusAlert{}).
			Where("alert_id = ?", alertId).
			Updates(updates).Error
	}
}

// ResetDailyNotifyCount 重置每日通知计数(跨天时调用)
func (s *AlertDedupService) ResetDailyNotifyCount(alert *observe.PrometheusAlert) {
	now := time.Now()
	alert.DailyNotifyCount = 1
	alert.LastNotifyDate = &now
}
