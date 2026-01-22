package observe

// MQMessageRequest MQ消息请求结构
type MQMessageRequest struct {
	Topic string      `json:"topic"`
	Tag   string      `json:"tag"`
	Data  MQAlertData `json:"data"`
}

// MQAlertData MQ告警数据
type MQAlertData struct {
	Title       string        `json:"title"`    // 原 邮件主题
	Receiver    string        `json:"receiver"` // 新增
	AlertDetail MQAlertDetail `json:"detail"`   // 原 告警详情
}

// MQAlertDetail 告警详情
type MQAlertDetail struct {
	Status       string `json:"status"`       // 原 告警状态
	Severity     string `json:"severity"`     // 原 告警等级
	Cluster      string `json:"cluster"`      // 原 告警集群
	Object       string `json:"object"`       // 原 告警对象
	Indicator    string `json:"indicator"`    // 原 策略名称
	Summary      string `json:"summary"`      // 原 告警描述
	TriggerValue string `json:"triggerValue"` // 原 触发数值
	AlertTime    string `json:"alertTime"`    // 原 告警时间
	Remark       string `json:"remark"`       // 新增
}
