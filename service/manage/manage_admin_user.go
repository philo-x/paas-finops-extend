package manage

import (
	"errors"
	"gorm.io/gorm"
	"main.go/global"
	"main.go/model/manage"
	manageReq "main.go/model/manage/request"
	"main.go/utils"
	"strconv"
	"strings"
	"time"
)

type ManageAdminUserService struct {
}

// CreateFinopsAdminUser 创建FinopsAdminUser记录
func (m *ManageAdminUserService) CreateFinopsAdminUser(finopsAdminUser manage.FinopsAdminUser) (err error) {
	if !errors.Is(global.GVA_DB.Where("login_user_name = ?", finopsAdminUser.LoginUserName).First(&manage.FinopsAdminUser{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同用户名")
	}
	err = global.GVA_DB.Create(&finopsAdminUser).Error
	return err
}

// UpdateFinopsAdminName 更新FinopsAdminUser昵称
func (m *ManageAdminUserService) UpdateFinopsAdminName(token string, req manageReq.FinopsUpdateNameParam) (err error) {
	var adminUserToken manage.FinopsAdminUserToken
	err = global.GVA_DB.Where("token =? ", token).First(&adminUserToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	err = global.GVA_DB.Where("admin_user_id = ?", adminUserToken.AdminUserId).Updates(&manage.FinopsAdminUser{
		LoginUserName: req.LoginUserName,
		NickName:      req.NickName,
	}).Error
	return err
}

func (m *ManageAdminUserService) UpdateFinopsAdminPassWord(token string, req manageReq.FinopsUpdatePasswordParam) (err error) {
	var adminUserToken manage.FinopsAdminUserToken
	err = global.GVA_DB.Where("token =? ", token).First(&adminUserToken).Error
	if err != nil {
		return errors.New("用户未登录")
	}
	var adminUser manage.FinopsAdminUser
	err = global.GVA_DB.Where("admin_user_id =?", adminUserToken.AdminUserId).First(&adminUser).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	if adminUser.LoginPassword != req.OriginalPassword {
		return errors.New("原密码不正确")
	}
	adminUser.LoginPassword = req.NewPassword

	err = global.GVA_DB.Where("admin_user_id=?", adminUser.AdminUserId).Updates(&adminUser).Error
	return
}

// GetFinopsAdminUser 根据id获取FinopsAdminUser记录
func (m *ManageAdminUserService) GetFinopsAdminUser(token string) (err error, finopsAdminUser manage.FinopsAdminUser) {
	var adminToken manage.FinopsAdminUserToken
	if errors.Is(global.GVA_DB.Where("token =?", token).First(&adminToken).Error, gorm.ErrRecordNotFound) {
		return errors.New("不存在的用户"), finopsAdminUser
	}
	err = global.GVA_DB.Where("admin_user_id = ?", adminToken.AdminUserId).First(&finopsAdminUser).Error
	return err, finopsAdminUser
}

// AdminLogin 管理员登陆
func (m *ManageAdminUserService) AdminLogin(params manageReq.FinopsAdminLoginParam) (err error, finopsAdminUser manage.FinopsAdminUser, adminToken manage.FinopsAdminUserToken) {
	err = global.GVA_DB.Where("login_user_name=? AND login_password=?", params.UserName, params.PasswordMd5).First(&finopsAdminUser).Error
	if finopsAdminUser != (manage.FinopsAdminUser{}) {
		token := getNewToken(time.Now().UnixNano()/1e6, int(finopsAdminUser.AdminUserId))
		global.GVA_DB.Where("admin_user_id", finopsAdminUser.AdminUserId).First(&adminToken)
		nowDate := time.Now()
		// 48小时过期
		expireTime, _ := time.ParseDuration("48h")
		expireDate := nowDate.Add(expireTime)
		// 没有token新增，有token 则更新
		if adminToken == (manage.FinopsAdminUserToken{}) {
			adminToken.AdminUserId = finopsAdminUser.AdminUserId
			adminToken.Token = token
			adminToken.UpdateTime = nowDate
			adminToken.ExpireTime = expireDate
			if err = global.GVA_DB.Create(&adminToken).Error; err != nil {
				return
			}
		} else {
			adminToken.Token = token
			adminToken.UpdateTime = nowDate
			adminToken.ExpireTime = expireDate
			if err = global.GVA_DB.Save(&adminToken).Error; err != nil {
				return
			}
		}
	}
	return err, finopsAdminUser, adminToken

}

func getNewToken(timeInt int64, userId int) (token string) {
	var build strings.Builder
	build.WriteString(strconv.FormatInt(timeInt, 10))
	build.WriteString(strconv.Itoa(userId))
	build.WriteString(utils.GenValidateCode(6))
	return utils.MD5V([]byte(build.String()))
}
