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
