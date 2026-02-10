package observe

import (
	"crypto/md5"
	"fmt"
	"time"

	"gorm.io/gorm"
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
// 发送成功后调用此方法: 清除NotifyPending标记
// 注意: DailyNotifyCount已在TryReserveNotification中原子预占
func (s *AlertDedupService) ConfirmNotifySent(alertId int) error {
	return global.GVA_DB.Model(&observe.PrometheusAlert{}).
		Where("alert_id = ?", alertId).
		Update("notify_pending", false).Error
}

// TryReserveNotification 尝试原子预占通知配额
// 返回 true 表示成功预占，可以发送通知
// 使用乐观锁：UPDATE ... WHERE daily_notify_count < limit
func (s *AlertDedupService) TryReserveNotification(alertId int) (bool, error) {
	dailyLimit := global.GVA_CONFIG.MQ.DailyNotifyLimit
	if dailyLimit <= 0 {
		// 不限制，直接返回成功
		return true, nil
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// 原子操作：检查并预占配额
	// 条件：同一天且未达上限，或者跨天（重置计数）
	result := global.GVA_DB.Model(&observe.PrometheusAlert{}).
		Where("alert_id = ?", alertId).
		Where("(DATE(last_notify_date) = ? AND daily_notify_count < ?) OR last_notify_date IS NULL OR DATE(last_notify_date) != ?",
			today, dailyLimit, today).
		Updates(map[string]interface{}{
			"notify_pending":     true,
			"daily_notify_count": gorm.Expr("CASE WHEN DATE(last_notify_date) = ? THEN daily_notify_count + 1 ELSE 1 END", today),
			"last_notify_date":   now,
		})

	if result.Error != nil {
		return false, result.Error
	}

	// RowsAffected > 0 表示成功预占
	return result.RowsAffected > 0, nil
}

// RollbackNotification 回滚通知计数（发送失败时调用）
func (s *AlertDedupService) RollbackNotification(alertId int) error {
	return global.GVA_DB.Model(&observe.PrometheusAlert{}).
		Where("alert_id = ?", alertId).
		UpdateColumn("daily_notify_count", gorm.Expr("GREATEST(daily_notify_count - 1, 0)")).Error
}

// ResetDailyNotifyCount 重置每日通知计数(跨天时调用)
func (s *AlertDedupService) ResetDailyNotifyCount(alert *observe.PrometheusAlert) {
	now := time.Now()
	alert.DailyNotifyCount = 1
	alert.LastNotifyDate = &now
}
