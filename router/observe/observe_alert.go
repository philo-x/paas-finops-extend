package observe

import (
	"github.com/gin-gonic/gin"
	v1 "main.go/api/v1"
)

type ObserveAlertRouter struct {
}

func (r *ObserveAlertRouter) InitObserveAlertRouter(Router *gin.RouterGroup) {
	alertRouter := Router
	var alertApi = v1.ApiGroupApp.ObserveApiGroup.ObserveAlertApi
	{
		alertRouter.POST("alerts", alertApi.CreateAlert)
		alertRouter.DELETE("alerts/:alertId", alertApi.DeleteAlert)
		alertRouter.DELETE("alerts", alertApi.DeleteAlertBatch)
		alertRouter.PUT("alerts/:alertId", alertApi.UpdateAlert)
		alertRouter.GET("alerts/:alertId", alertApi.GetAlert)
		alertRouter.GET("alerts", alertApi.GetAlertList)
	}
}
