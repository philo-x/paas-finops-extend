package request

type FinopsAdminLoginParam struct {
	UserName    string `json:"userName"`
	PasswordMd5 string `json:"passwordMd5"`
}

type FinopsAdminParam struct {
	LoginUserName string `json:"loginUserName"`
	LoginPassword string `json:"loginPassword"`
	NickName      string `json:"nickName"`
}

type FinopsUpdateNameParam struct {
	LoginUserName string `json:"loginUserName"`
	NickName      string `json:"nickName"`
}

type FinopsUpdatePasswordParam struct {
	OriginalPassword string `json:"originalPassword"`
	NewPassword      string `json:"newPassword"`
}
