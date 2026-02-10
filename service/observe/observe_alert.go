package observe

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/common/request"
	"main.go/model/observe"
)

type ObserveAlertService struct {
}

// CreateAlert 创建告警(含去重和通知限流逻辑)
// 使用 Upsert 模式解决并发问题：通过数据库唯一约束保证同一指纹的告警只有一条记录
func (m *ObserveAlertService) CreateAlert(req observe.AlertRequest) (err error, alert observe.PrometheusAlert) {
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return err, alert
	}

	var endsAt time.Time
	if req.EndsAt != "" {
		endsAt, err = time.Parse(time.RFC3339, req.EndsAt)
		if err != nil {
			return err, alert
		}
	}

	// 生成告警指纹（不包含状态，以便firing和resolved可以匹配）
	dedupService := AlertDedupService{}
	fingerprint := dedupService.GenerateFingerprint(req.Labels)
	now := time.Now()

	// 构建告警对象
	alert = observe.PrometheusAlert{
		Status:           req.Status,
		StartsAt:         &observe.NullTime{Time: &startsAt},
		EndsAt:           &observe.NullTime{Time: &endsAt},
		Annotations:      req.Annotations,
		Labels:           req.Labels,
		Fingerprint:      fingerprint,
		AlertCount:       1,
		DailyNotifyCount: 0,
		NotifyPending:    true, // 新告警标记为待发送
		LastNotifyDate:   nil,
		IsDeleted:        0,
		CreateTime:       common.JSONTime{Time: now},
		UpdateTime:       common.JSONTime{Time: now},
	}

	// 原子 Upsert: 插入或更新
	// 当 fingerprint+is_deleted 冲突时，更新现有记录并增加 alert_count
	err = global.GVA_DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "fingerprint"}, {Name: "is_deleted"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":      req.Status,
			"starts_at":   startsAt,
			"ends_at":     endsAt,
			"annotations": req.Annotations,
			"labels":      req.Labels,
			"alert_count": gorm.Expr("alert_count + 1"),
			"update_time": now,
		}),
	}).Create(&alert).Error

	if err != nil {
		return err, alert
	}

	// 获取最新记录（upsert 后需要获取完整数据，包括正确的 alert_id 和 alert_count）
	err = global.GVA_DB.Where("fingerprint = ? AND is_deleted = 0", fingerprint).First(&alert).Error
	if err != nil {
		return err, alert
	}

	// 判断是否需要发送通知
	// 始终使用乐观锁原子预占通知配额，解决并发竞态问题
	reserved, reserveErr := dedupService.TryReserveNotification(alert.AlertId)
	if reserveErr != nil {
		global.GVA_LOG.Error("预占通知配额失败", zap.Error(reserveErr), zap.Int("alertId", alert.AlertId))
	}
	shouldNotify := reserved

	// 异步发送MQ通知(失败仅记录日志，不影响主流程)
	if shouldNotify {
		go func(alertId int, alertCopy observe.PrometheusAlert) {
			mqService := MQClientService{}
			if sendErr := mqService.SendAlertNotification(alertCopy); sendErr != nil {
				global.GVA_LOG.Error("MQ通知发送失败", zap.Error(sendErr), zap.Int("alertId", alertId))
				// 发送失败时回滚计数
				if rollbackErr := dedupService.RollbackNotification(alertId); rollbackErr != nil {
					global.GVA_LOG.Error("回滚通知计数失败", zap.Error(rollbackErr), zap.Int("alertId", alertId))
				}
			} else {
				// 发送成功，确认通知已发送(清除NotifyPending)
				if confirmErr := dedupService.ConfirmNotifySent(alertId); confirmErr != nil {
					global.GVA_LOG.Error("确认通知发送状态失败", zap.Error(confirmErr), zap.Int("alertId", alertId))
				} else {
					global.GVA_LOG.Info("MQ通知发送成功", zap.Int("alertId", alertId))
				}
			}
		}(alert.AlertId, alert)
	} else {
		global.GVA_LOG.Info("跳过MQ通知(已达每日限制)",
			zap.Int("alertId", alert.AlertId),
			zap.String("fingerprint", fingerprint),
			zap.Int("alertCount", alert.AlertCount),
			zap.Int("dailyNotifyCount", alert.DailyNotifyCount),
		)
	}

	return nil, alert
}

// DeleteAlert 删除告警（软删除）
func (m *ObserveAlertService) DeleteAlert(id int) (err error) {
	err = global.GVA_DB.Model(&observe.PrometheusAlert{}).Where("alert_id = ?", id).Updates(map[string]interface{}{
		"is_deleted":  1,
		"update_time": common.JSONTime{Time: time.Now()},
	}).Error
	return err
}

// DeleteAlertBatch 批量删除告警
func (m *ObserveAlertService) DeleteAlertBatch(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Model(&observe.PrometheusAlert{}).Where("alert_id in ?", ids.Ids).Updates(map[string]interface{}{
		"is_deleted":  1,
		"update_time": common.JSONTime{Time: time.Now()},
	}).Error
	return err
}

// UpdateAlert 更新告警
func (m *ObserveAlertService) UpdateAlert(id int, req observe.AlertRequest) (err error) {
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return err
	}

	var endsAt time.Time
	if req.EndsAt != "" {
		endsAt, err = time.Parse(time.RFC3339, req.EndsAt)
		if err != nil {
			return err
		}
	}

	err = global.GVA_DB.Model(&observe.PrometheusAlert{}).Where("alert_id = ? AND is_deleted = 0", id).Updates(map[string]interface{}{
		"status":      req.Status,
		"starts_at":   startsAt,
		"ends_at":     endsAt,
		"annotations": req.Annotations,
		"labels":      req.Labels,
		"update_time": common.JSONTime{Time: time.Now()},
	}).Error
	return err
}

// GetAlert 根据ID获取告警
func (m *ObserveAlertService) GetAlert(id int) (err error, alert observe.PrometheusAlert) {
	err = global.GVA_DB.Where("alert_id = ? AND is_deleted = 0", id).First(&alert).Error
	return err, alert
}

// GetAlertList 分页获取告警列表
func (m *ObserveAlertService) GetAlertList(info request.PageInfo, status string, severity string) (err error, list []observe.PrometheusAlert, total int64) {
	limit := info.PageSize
	if limit == 0 {
		limit = 10
	}
	offset := limit * (info.PageNumber - 1)
	if info.PageNumber == 0 {
		offset = 0
	}

	db := global.GVA_DB.Model(&observe.PrometheusAlert{}).Where("is_deleted = 0")

	if status != "" {
		db = db.Where("status = ?", status)
	}

	if severity != "" {
		db = db.Where("JSON_EXTRACT(labels, '$.severity') = ?", severity)
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	err = db.Limit(limit).Offset(offset).Order("create_time desc").Find(&list).Error
	return err, list, total
}
