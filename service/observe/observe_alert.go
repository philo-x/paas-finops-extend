package observe

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/common/request"
	"main.go/model/observe"
)

type ObserveAlertService struct {
}

// CreateAlert 创建告警(含去重和通知限流逻辑)
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

	// 生成告警指纹
	dedupService := AlertDedupService{}
	fingerprint := dedupService.GenerateFingerprint(req.Labels)

	isResolved := req.Status == "resolved"
	shouldNotify := false
	today := time.Now().Truncate(24 * time.Hour)

	// 查找是否存在相同指纹的活跃告警
	existingAlert, findErr := dedupService.FindActiveAlertByFingerprint(fingerprint)

	if findErr == nil && existingAlert != nil {
		// 存在相同指纹的活跃告警，更新现有记录
		existingAlert.AlertCount++
		existingAlert.Status = req.Status
		existingAlert.StartsAt = &observe.NullTime{Time: &startsAt}
		if req.EndsAt != "" {
			existingAlert.EndsAt = &observe.NullTime{Time: &endsAt}
		}
		existingAlert.Annotations = req.Annotations
		existingAlert.Labels = req.Labels
		existingAlert.UpdateTime = common.JSONTime{Time: time.Now()}

		// 判断是否需要发送通知
		shouldNotify = dedupService.ShouldSendNotification(existingAlert, isResolved)

		if shouldNotify && !isResolved {
			// firing状态且需要发送通知，更新通知计数
			dedupService.UpdateNotifyCount(existingAlert)
		}

		// 更新数据库
		err = global.GVA_DB.Save(existingAlert).Error
		if err != nil {
			return err, alert
		}
		alert = *existingAlert

	} else if findErr == gorm.ErrRecordNotFound || existingAlert == nil {
		// 不存在相同指纹的活跃告警，创建新记录
		alert = observe.PrometheusAlert{
			Status:           req.Status,
			StartsAt:         &observe.NullTime{Time: &startsAt},
			EndsAt:           &observe.NullTime{Time: &endsAt},
			Annotations:      req.Annotations,
			Labels:           req.Labels,
			Fingerprint:      fingerprint,
			AlertCount:       1,
			DailyNotifyCount: 1,
			LastNotifyDate:   &today,
			IsDeleted:        0,
			CreateTime:       common.JSONTime{Time: time.Now()},
			UpdateTime:       common.JSONTime{Time: time.Now()},
		}

		err = global.GVA_DB.Create(&alert).Error
		if err != nil {
			return err, alert
		}

		// 新告警始终发送通知
		shouldNotify = true

	} else {
		// 查询出错
		return findErr, alert
	}

	// 异步发送MQ通知(失败仅记录日志，不影响主流程)
	if shouldNotify {
		go func() {
			mqService := MQClientService{}
			if sendErr := mqService.SendAlertNotification(alert); sendErr != nil {
				global.GVA_LOG.Error("MQ通知发送失败", zap.Error(sendErr), zap.Int("alertId", alert.AlertId))
			} else {
				global.GVA_LOG.Info("MQ通知发送成功", zap.Int("alertId", alert.AlertId))
			}
		}()
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
