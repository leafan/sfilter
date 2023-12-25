package models

type RegisterForm struct {
	Username  string `json:"username" bson:"username" binding:"required"`
	Passwd    string `json:"passwd" bson:"passwd" binding:"required"`
	ReferCode string `json:"refercode" bson:"refercode"`
	AuthCode  string `json:"authcode" bson:"authcode" binding:"required"`
}

type ResetPasswdForm struct {
	PasswdOld string `json:"passwdOld" binding:"required"`
	PasswdNew string `json:"passwdNew" binding:"required"`
}

// 新增跟踪地址
type TrackAddressForm struct {
	Address  string `json:"address" binding:"required"`
	Memo     string `json:"memo" binding:"required"`
	Priority int    `json:"priority"`
}

// 修改跟踪地址
type UpdateTrackAddressForm struct {
	Memo     string `json:"memo" binding:"required"`
	Priority int    `json:"priority"`
}

// 修改用户role
type AdminUpdateUserRoleForm struct {
	Username string `json:"username" binding:"required"`
	Role     int    `json:"role"  binding:"required"`
}
