package middleware

import (
	"github.com/gin-gonic/gin"
	"main.go/model/common/response"
	"main.go/service"
	"time"
)

var manageAdminUserTokenService = service.ServiceGroupApp.ManageServiceGroup.ManageAdminUserTokenService

func AdminJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if token == "" {
			response.FailWithDetailed(nil, "未登录或非法访问", c)
			c.Abort()
			return
		}
		err, finopsAdminUserToken := manageAdminUserTokenService.ExistAdminToken(token)
		if err != nil {
			response.FailWithDetailed(nil, "未登录或非法访问", c)
			c.Abort()
			return
		}
		if time.Now().After(finopsAdminUserToken.ExpireTime) {
			response.FailWithDetailed(nil, "授权已过期", c)
			err = manageAdminUserTokenService.DeleteAdminUserToken(token)
			if err != nil {
				return
			}
			c.Abort()
			return
		}
		c.Next()
	}
}
