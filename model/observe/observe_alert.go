package observe

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"main.go/model/common"
)

// AlertAnnotations 告警注解
type AlertAnnotations struct {
	AlertCurrentValue  string `json:"alert_current_value"`
	AlertNotifications string `json:"alert_notifications"` // JSON数组字符串 "[\"target1\",\"target2\"]"
	DisplayName        string `json:"display_name"`        // JSON字符串 {"zh":"...", "en":"..."}
	Summary            string `json:"summary"`             // JSON字符串 {"zh":"...", "en":"..."}
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
	AlertIndicatorAlias          string `json:"alert_indicator_alias"`
	AlertIndicatorComparison     string `json:"alert_indicator_comparison"`
	AlertIndicatorThreshold      string `json:"alert_indicator_threshold"`
	AlertInvolvedObjectKind      string `json:"alert_involved_object_kind"`
	AlertInvolvedObjectName      string `json:"alert_involved_object_name"`
	AlertInvolvedObjectOptions   string `json:"alert_involved_object_options"`
	AlertKind                    string `json:"alert_kind"`
	AlertName                    string `json:"alert_name"`
	AlertNamespace               string `json:"alert_namespace"`
	AlertProject                 string `json:"alert_project"`
	AlertResource                string `json:"alert_resource"`
	AlertSource                  string `json:"alert_source"`
	Alertname                    string `json:"alertname"`
	DisplayName                  string `json:"display_name"`
	NodeName                     string `json:"node_name"`
	Severity                     string `json:"severity"`
	// 新增字段 - 兼容多种数据源格式
	HostIP             string `json:"host_ip"`
	Instance           string `json:"instance"`
	IP                 string `json:"ip"`
	Node               string `json:"node"`
	Device             string `json:"device"`
	Mountpoint         string `json:"mountpoint"`
	AlertIndicatorUnit string `json:"alert_indicator_unit"`
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
	AlertId          int              `json:"alertId" form:"alertId" gorm:"primarykey;AUTO_INCREMENT"`
	Status           string           `json:"status" form:"status" gorm:"column:status;comment:告警状态;type:varchar(50);"`
	StartsAt         *NullTime        `json:"startsAt" form:"startsAt" gorm:"column:starts_at;comment:告警开始时间;type:datetime;"`
	EndsAt           *NullTime        `json:"endsAt" form:"endsAt" gorm:"column:ends_at;comment:告警结束时间;type:datetime;"`
	Annotations      AlertAnnotations `json:"annotations" form:"annotations" gorm:"column:annotations;comment:告警注解;type:json;"`
	Labels           AlertLabels      `json:"labels" form:"labels" gorm:"column:labels;comment:告警标签;type:json;"`
	Fingerprint      string           `json:"fingerprint" form:"fingerprint" gorm:"column:fingerprint;comment:告警指纹;type:varchar(64);index"`
	AlertCount       int              `json:"alertCount" form:"alertCount" gorm:"column:alert_count;comment:累计告警次数;type:int;default:1"`
	DailyNotifyCount int              `json:"dailyNotifyCount" form:"dailyNotifyCount" gorm:"column:daily_notify_count;comment:当日通知次数;type:int;default:0"`
	LastNotifyDate   *time.Time       `json:"lastNotifyDate" form:"lastNotifyDate" gorm:"column:last_notify_date;comment:最后通知日期;type:date;"`
	NotifyPending    bool             `json:"notifyPending" form:"notifyPending" gorm:"column:notify_pending;comment:是否有待发送的通知;type:tinyint(1);default:0"`
	IsDeleted        int              `json:"isDeleted" form:"isDeleted" gorm:"column:is_deleted;comment:删除标识字段(0-未删除 1-已删除);type:tinyint;default:0"`
	CreateTime       common.JSONTime  `json:"createTime" form:"createTime" gorm:"column:create_time;comment:创建时间;type:datetime;"`
	UpdateTime       common.JSONTime  `json:"updateTime" form:"updateTime" gorm:"column:update_time;comment:最新修改时间;type:datetime;"`
}

// TableName PrometheusAlert 表名
func (PrometheusAlert) TableName() string {
	return "prometheus_alert"
}
