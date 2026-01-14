package observe

// MQMessageRequest MQ消息请求结构
type MQMessageRequest struct {
	Topic string      `json:"topic"`
	Tag   string      `json:"tag"`
	Data  MQAlertData `json:"data"`
}

// MQAlertData MQ告警数据
type MQAlertData struct {
	EmailSubject string        `json:"邮件主题"`
	AlertDetail  MQAlertDetail `json:"告警详情"`
}

// MQAlertDetail 告警详情
type MQAlertDetail struct {
	AlertStatus  string `json:"告警状态"`
	AlertLevel   string `json:"告警等级"`
	AlertCluster string `json:"告警集群"`
	AlertObject  string `json:"告警对象"`
	PolicyName   string `json:"策略名称"`
	AlertDesc    string `json:"告警描述"`
	TriggerValue string `json:"触发数值"`
	AlertTime    string `json:"告警时间"`
}
