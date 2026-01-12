package manage

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"main.go/global"
	"main.go/model/common/response"
	"main.go/model/example"
	"main.go/model/manage"
	manageReq "main.go/model/manage/request"
	"main.go/utils"
)

type ManageAdminUserApi struct {
}

// 创建AdminUser
func (m *ManageAdminUserApi) CreateAdminUser(c *gin.Context) {
	var params manageReq.FinopsAdminParam
	_ = c.ShouldBindJSON(&params)
	if err := utils.Verify(params, utils.AdminUserRegisterVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	finopsAdminUser := manage.FinopsAdminUser{
		LoginUserName: params.LoginUserName,
		NickName:      params.NickName,
		LoginPassword: utils.MD5V([]byte(params.LoginPassword)),
	}
	if err := finopsAdminUserService.CreateFinopsAdminUser(finopsAdminUser); err != nil {
		global.GVA_LOG.Error("创建失败:", zap.Error(err))
		response.FailWithMessage("创建失败"+err.Error(), c)
	} else {
		response.OkWithMessage("创建成功", c)
	}
}

// 修改密码
func (m *ManageAdminUserApi) UpdateAdminUserPassword(c *gin.Context) {
	var req manageReq.FinopsUpdatePasswordParam
	_ = c.ShouldBindJSON(&req)
	userToken := c.GetHeader("token")
	if err := finopsAdminUserService.UpdateFinopsAdminPassWord(userToken, req); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
	} else {
		response.OkWithMessage("更新成功", c)
	}

}

// 更新用户名
func (m *ManageAdminUserApi) UpdateAdminUserName(c *gin.Context) {
	var req manageReq.FinopsUpdateNameParam
	_ = c.ShouldBindJSON(&req)
	userToken := c.GetHeader("token")
	if err := finopsAdminUserService.UpdateFinopsAdminName(userToken, req); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
	} else {
		response.OkWithMessage("更新成功", c)
	}
}

// AdminUserProfile 用id查询AdminUser
func (m *ManageAdminUserApi) AdminUserProfile(c *gin.Context) {
	adminToken := c.GetHeader("token")
	if err, finopsAdminUser := finopsAdminUserService.GetFinopsAdminUser(adminToken); err != nil {
		global.GVA_LOG.Error("未查询到记录", zap.Error(err))
		response.FailWithMessage("未查询到记录", c)
	} else {
		finopsAdminUser.LoginPassword = "******"
		response.OkWithData(finopsAdminUser, c)
	}
}

// AdminLogin 管理员登陆
func (m *ManageAdminUserApi) AdminLogin(c *gin.Context) {
	var adminLoginParams manageReq.FinopsAdminLoginParam
	_ = c.ShouldBindJSON(&adminLoginParams)
	if err, _, adminToken := finopsAdminUserService.AdminLogin(adminLoginParams); err != nil {
		response.FailWithMessage("登陆失败", c)
	} else {
		response.OkWithData(adminToken.Token, c)
	}
}

// AdminLogout 登出
func (m *ManageAdminUserApi) AdminLogout(c *gin.Context) {
	token := c.GetHeader("token")
	if err := finopsAdminUserTokenService.DeleteFinopsAdminUserToken(token); err != nil {
		response.FailWithMessage("登出失败", c)
	} else {
		response.OkWithMessage("登出成功", c)
	}

}

// UploadFile 上传单图
func (m *ManageAdminUserApi) UploadFile(c *gin.Context) {
	var file example.ExaFileUploadAndDownload
	noSave := c.DefaultQuery("noSave", "0")
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		global.GVA_LOG.Error("接收文件失败!", zap.Error(err))
		response.FailWithMessage("接收文件失败", c)
		return
	}
	err, file = fileUploadAndDownloadService.UploadFile(header, noSave)
	if err != nil {
		global.GVA_LOG.Error("修改数据库链接失败!", zap.Error(err))
		response.FailWithMessage("修改数据库链接失败", c)
		return
	}
	response.OkWithData("http://localhost:8888/"+file.Url, c)
}
