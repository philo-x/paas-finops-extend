package initialize

import (
	"main.go/service"
)

func K8s() {
	service.ServiceGroupApp.WebhookServiceGroup.RecommendationService.InitK8s()
}
