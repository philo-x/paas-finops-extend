package manage

import (
	"github.com/gin-gonic/gin"
	v1 "main.go/api/v1"
	"main.go/middleware"
)

type ManageAdminUserRouter struct {
}

func (r *ManageAdminUserRouter) InitManageAdminUserRouter(Router *gin.RouterGroup) {
	adminUserRouter := Router.Use(middleware.AdminJWTAuth())
	adminUserWithoutRouter := Router
	var adminUserApi = v1.ApiGroupApp.ManageApiGroup.ManageAdminUserApi
	{
		adminUserRouter.POST("createAdminUser", adminUserApi.CreateAdminUser)
		adminUserRouter.PUT("adminUser/name", adminUserApi.UpdateAdminUserName)
		adminUserRouter.PUT("adminUser/password", adminUserApi.UpdateAdminUserPassword)
		adminUserRouter.GET("adminUser/profile", adminUserApi.AdminUserProfile)
		adminUserRouter.DELETE("logout", adminUserApi.AdminLogout)
		adminUserRouter.POST("upload/file", adminUserApi.UploadFile)
	}
	{
		adminUserWithoutRouter.POST("adminUser/login", adminUserApi.AdminLogin)
	}
}
