package webhook

import (
	"github.com/gin-gonic/gin"
	v1 "main.go/api/v1"
)

type WebhookRouter struct{}

func (s *WebhookRouter) InitWebhookRouter(Router *gin.RouterGroup) {
	webhookApi := v1.ApiGroupApp.WebhookApiGroup.RecommendationApi
	Router.POST("mutate", webhookApi.ServeMutate)
}
