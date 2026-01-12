package manage

import (
	"github.com/gin-gonic/gin"
	v1 "main.go/api/v1"
	"main.go/middleware"
)

type ManageAdminUserRouter struct {
}

func (r *ManageAdminUserRouter) InitManageAdminUserRouter(Router *gin.RouterGroup) {
	finopsAdminUserRouter := Router.Group("v1").Use(middleware.AdminJWTAuth())
	finopsAdminUserWithoutRouter := Router.Group("v1")
	var finopsAdminUserApi = v1.ApiGroupApp.ManageApiGroup.ManageAdminUserApi
	{
		finopsAdminUserRouter.POST("createFinopsAdminUser", finopsAdminUserApi.CreateAdminUser)
		finopsAdminUserRouter.PUT("adminUser/name", finopsAdminUserApi.UpdateAdminUserName)
		finopsAdminUserRouter.PUT("adminUser/password", finopsAdminUserApi.UpdateAdminUserPassword)
		finopsAdminUserRouter.GET("adminUser/profile", finopsAdminUserApi.AdminUserProfile)
		finopsAdminUserRouter.DELETE("logout", finopsAdminUserApi.AdminLogout)
		finopsAdminUserRouter.POST("upload/file", finopsAdminUserApi.UploadFile)
	}
	{
		finopsAdminUserWithoutRouter.POST("adminUser/login", finopsAdminUserApi.AdminLogin)
	}
}
