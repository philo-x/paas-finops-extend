package utils

var (
	AdminUserRegisterVerify       = Rules{"Username": {NotEmpty()}, "NickName": {NotEmpty()}, "Password": {NotEmpty()}}
	AdminUserChangePasswordVerify = Rules{"Password": {NotEmpty()}}
)
