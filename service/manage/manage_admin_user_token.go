package manage

import (
	"main.go/global"
	"main.go/model/manage"
)

type ManageAdminUserTokenService struct {
}

func (m *ManageAdminUserTokenService) ExistAdminToken(token string) (err error, finopsAdminUserToken manage.FinopsAdminUserToken) {
	err = global.GVA_DB.Where("token =?", token).First(&finopsAdminUserToken).Error
	return
}

func (m *ManageAdminUserTokenService) DeleteFinopsAdminUserToken(token string) (err error) {
	err = global.GVA_DB.Delete(&[]manage.FinopsAdminUserToken{}, "token =?", token).Error
	return err
}
