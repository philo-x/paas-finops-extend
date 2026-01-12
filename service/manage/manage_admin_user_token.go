package manage

import (
	"main.go/global"
	"main.go/model/manage"
)

type ManageAdminUserTokenService struct {
}

func (m *ManageAdminUserTokenService) ExistAdminToken(token string) (err error, finopsAdminUserToken manage.AdminUserToken) {
	err = global.GVA_DB.Where("token =?", token).First(&finopsAdminUserToken).Error
	return
}

func (m *ManageAdminUserTokenService) DeleteAdminUserToken(token string) (err error) {
	err = global.GVA_DB.Where("token = ?", token).Delete(&manage.AdminUserToken{}).Error
	return err
}
