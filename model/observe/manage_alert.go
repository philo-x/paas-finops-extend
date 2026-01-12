package observe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"main.go/model/common"
)

// AlertNotification 告警通知
type AlertNotification struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// AlertAnnotations 告警注解
type AlertAnnotations struct {
	AlertCurrentValue  string              `json:"alert_current_value"`
	AlertNotifications []AlertNotification `json:"alert_notifications"`
}

// Value 实现 driver.Valuer 接口
func (a AlertAnnotations) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *AlertAnnotations) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// AlertLabels 告警标签
type AlertLabels struct {
	AlertCluster                 string `json:"alert_cluster"`
	AlertIndicator               string `json:"alert_indicator"`
	AlertIndicatorAggregateRange string `json:"alert_indicator_aggregate_range"`
	AlertIndicatorComparison     string `json:"alert_indicator_comparison"`
	AlertIndicatorThreshold      string `json:"alert_indicator_threshold"`
	AlertInvolvedObjectKind      string `json:"alert_involved_object_kind"`
	AlertInvolvedObjectName      string `json:"alert_involved_object_name"`
	AlertName                    string `json:"alert_name"`
	AlertResource                string `json:"alert_resource"`
	AlertName2                   string `json:"alertname"`
	Severity                     string `json:"severity"`
}

// Value 实现 driver.Valuer 接口
func (l AlertLabels) Value() (driver.Value, error) {
	return json.Marshal(l)
}

// Scan 实现 sql.Scanner 接口
func (l *AlertLabels) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, l)
}

// PrometheusAlert 告警信息模型
type PrometheusAlert struct {
	AlertId     int              `json:"alertId" form:"alertId" gorm:"primarykey;AUTO_INCREMENT"`
	Status      string           `json:"status" form:"status" gorm:"column:status;comment:告警状态;type:varchar(50);"`
	StartsAt    time.Time        `json:"startsAt" form:"startsAt" gorm:"column:starts_at;comment:告警开始时间;type:datetime;"`
	EndsAt      time.Time        `json:"endsAt" form:"endsAt" gorm:"column:ends_at;comment:告警结束时间;type:datetime;"`
	Annotations AlertAnnotations `json:"annotations" form:"annotations" gorm:"column:annotations;comment:告警注解;type:json;"`
	Labels      AlertLabels      `json:"labels" form:"labels" gorm:"column:labels;comment:告警标签;type:json;"`
	IsDeleted   int              `json:"isDeleted" form:"isDeleted" gorm:"column:is_deleted;comment:删除标识字段(0-未删除 1-已删除);type:tinyint;default:0"`
	CreateTime  common.JSONTime  `json:"createTime" form:"createTime" gorm:"column:create_time;comment:创建时间;type:datetime;"`
	UpdateTime  common.JSONTime  `json:"updateTime" form:"updateTime" gorm:"column:update_time;comment:最新修改时间;type:datetime;"`
}

// TableName PrometheusAlert 表名
func (PrometheusAlert) TableName() string {
	return "prometheus_alert"
}

// AlertRequest 创建/更新告警请求结构
type AlertRequest struct {
	Status      string           `json:"status" binding:"required"`
	StartsAt    string           `json:"startsAt" binding:"required"`
	EndsAt      string           `json:"endsAt"`
	Annotations AlertAnnotations `json:"annotations"`
	Labels      AlertLabels      `json:"labels"`
}
