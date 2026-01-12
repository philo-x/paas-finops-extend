package manage

import "time"

// FinopsAdminUserToken 管理员登录态
type FinopsAdminUserToken struct {
	AdminUserId int       `json:"adminUserId" form:"adminUserId" gorm:"primarykey;AUTO_INCREMENT"`
	Token       string    `json:"token" form:"token" gorm:"column:token;comment:token值(32位字符串);type:varchar(32);"`
	UpdateTime  time.Time `json:"updateTime" form:"updateTime" gorm:"column:update_time;comment:修改时间;type:datetime"`
	ExpireTime  time.Time `json:"expireTime" form:"expireTime" gorm:"column:expire_time;comment:token过期时间;type:datetime"`
}

func (FinopsAdminUserToken) TableName() string {
	return "finops_admin_user_token"
}
