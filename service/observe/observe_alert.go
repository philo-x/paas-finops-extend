package observe

import (
	"time"

	"go.uber.org/zap"
	"main.go/global"
	"main.go/model/common"
	"main.go/model/common/request"
	"main.go/model/observe"
)

type ObserveAlertService struct {
}

// CreateAlert 创建告警
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

	alert = observe.PrometheusAlert{
		Status:      req.Status,
		StartsAt:    &observe.NullTime{Time: &startsAt},
		EndsAt:      &observe.NullTime{Time: &endsAt},
		Annotations: req.Annotations,
		Labels:      req.Labels,
		IsDeleted:   0,
		CreateTime:  common.JSONTime{Time: time.Now()},
		UpdateTime:  common.JSONTime{Time: time.Now()},
	}

	err = global.GVA_DB.Create(&alert).Error
	if err != nil {
		return err, alert
	}

	// 异步发送MQ通知(失败仅记录日志，不影响主流程)
	go func() {
		mqService := MQClientService{}
		if sendErr := mqService.SendAlertNotification(alert); sendErr != nil {
			global.GVA_LOG.Error("MQ通知发送失败", zap.Error(sendErr), zap.Int("alertId", alert.AlertId))
		} else {
			global.GVA_LOG.Info("MQ通知发送成功", zap.Int("alertId", alert.AlertId))
		}
	}()

	return err, alert
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
