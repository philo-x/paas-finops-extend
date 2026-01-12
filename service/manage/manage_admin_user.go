package manage

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"main.go/global"
	"main.go/model/manage"
	manageReq "main.go/model/manage/request"
	"main.go/utils"
)

type ManageAdminUserService struct {
}

// CreateadminUser 创建adminUser记录
func (m *ManageAdminUserService) CreateAdminUser(adminUser manage.AdminUser) (err error) {
	if !errors.Is(global.GVA_DB.Where("login_user_name = ?", adminUser.LoginUserName).First(&manage.AdminUser{}).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同用户名")
	}
	err = global.GVA_DB.Create(&adminUser).Error
	return err
}

// UpdateadminName 更新adminUser昵称
func (m *ManageAdminUserService) UpdateAdminName(token string, req manageReq.UserNameUpdateParam) (err error) {
	var adminUserToken manage.AdminUserToken
	err = global.GVA_DB.Where("token =? ", token).First(&adminUserToken).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	err = global.GVA_DB.Where("admin_user_id = ?", adminUserToken.AdminUserId).Updates(&manage.AdminUser{
		LoginUserName: req.LoginUserName,
		NickName:      req.NickName,
	}).Error
	return err
}

func (m *ManageAdminUserService) UpdateAdminPassWord(token string, req manageReq.UserPasswordUpdateParam) (err error) {
	var adminUserToken manage.AdminUserToken
	err = global.GVA_DB.Where("token =? ", token).First(&adminUserToken).Error
	if err != nil {
		return errors.New("用户未登录")
	}
	var adminUser manage.AdminUser
	err = global.GVA_DB.Where("admin_user_id =?", adminUserToken.AdminUserId).First(&adminUser).Error
	if err != nil {
		return errors.New("不存在的用户")
	}
	if adminUser.LoginPassword != utils.MD5V([]byte(req.OriginalPassword)) {
		return errors.New("原密码不正确")
	}
	adminUser.LoginPassword = utils.MD5V([]byte(req.NewPassword))

	err = global.GVA_DB.Where("admin_user_id=?", adminUser.AdminUserId).Updates(&adminUser).Error
	return
}

// GetadminUser 根据id获取adminUser记录
func (m *ManageAdminUserService) GetAdminUser(token string) (err error, adminUser manage.AdminUser) {
	var adminToken manage.AdminUserToken
	if errors.Is(global.GVA_DB.Where("token =?", token).First(&adminToken).Error, gorm.ErrRecordNotFound) {
		return errors.New("不存在的用户"), adminUser
	}
	err = global.GVA_DB.Where("admin_user_id = ?", adminToken.AdminUserId).First(&adminUser).Error
	return err, adminUser
}

// AdminLogin 管理员登陆
func (m *ManageAdminUserService) AdminLogin(params manageReq.AdminLoginParam) (err error, adminUser manage.AdminUser, adminToken manage.AdminUserToken) {
	err = global.GVA_DB.Where("login_user_name=? AND login_password=?", params.UserName, params.PasswordMd5).First(&adminUser).Error
	if adminUser != (manage.AdminUser{}) {
		token := getNewToken(time.Now().UnixNano()/1e6, int(adminUser.AdminUserId))
		global.GVA_DB.Where("admin_user_id = ?", adminUser.AdminUserId).First(&adminToken)
		nowDate := time.Now()
		// 48小时过期
		expireTime, _ := time.ParseDuration("48h")
		expireDate := nowDate.Add(expireTime)
		// 没有token新增，有token 则更新
		if adminToken == (manage.AdminUserToken{}) {
			adminToken.AdminUserId = adminUser.AdminUserId
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
	return err, adminUser, adminToken

}

func getNewToken(timeInt int64, userId int) (token string) {
	var build strings.Builder
	build.WriteString(strconv.FormatInt(timeInt, 10))
	build.WriteString(strconv.Itoa(userId))
	build.WriteString(utils.GenValidateCode(6))
	return utils.MD5V([]byte(build.String()))
}
