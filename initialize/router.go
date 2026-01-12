package initialize

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"main.go/global"
	"main.go/middleware"
	"main.go/router"
)

func Routers() *gin.Engine {
	var Router = gin.Default()
	Router.StaticFS(global.GVA_CONFIG.Local.Path, http.Dir(global.GVA_CONFIG.Local.Path)) // 为用户头像和文件提供静态地址
	global.GVA_LOG.Info("use middleware logger")
	// 跨域
	Router.Use(middleware.Cors())
	global.GVA_LOG.Info("use middleware cors")

	// public 路由
	PublicGroup := Router.Group("")
	{
		// 健康监测
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})
	}

	// 管理路由
	manageRouter := router.RouterGroupApp.Manage
	ManageGroup := Router.Group("api/v1/manage")
	{
		// 管理路由初始化
		manageRouter.InitManageAdminUserRouter(ManageGroup)
	}

	// 告警路由
	observeRouter := router.RouterGroupApp.Observe
	AlertGroup := Router.Group("api/v1/observe")
	{
		// 告警路由初始化
		observeRouter.InitObserveAlertRouter(AlertGroup)
	}

	global.GVA_LOG.Info("router register success")
	return Router
}
