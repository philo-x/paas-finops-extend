package observe

// AlertRequest 创建/更新告警请求结构
type AlertRequest struct {
	Status      string           `json:"status" binding:"required"`
	StartsAt    string           `json:"startsAt" binding:"required"`
	EndsAt      string           `json:"endsAt"`
	Annotations AlertAnnotations `json:"annotations"`
	Labels      AlertLabels      `json:"labels"`
}
