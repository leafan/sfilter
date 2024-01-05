package controllers

import (
	"time"

	"sfilter/user/models"
	"sfilter/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-pkgz/auth/token"
)

func Register(c *gin.Context) {
	params, err := getRegisterParams(c)
	if err != nil {
		ResFailure(c, 401, err.Error())
		return
	}

	err = preCheckRegister(params)
	if err != nil {
		ResFailure(c, 401, err.Error())
		return
	}

	params.Passwd, err = utils.HashPassword(params.Passwd)
	if err != nil {
		ResFailure(c, 401, err.Error())
		return
	}

	role := models.USER_ROLE_BASIC

	user := &models.User{
		BasicInfo: models.BasicInfo{
			Username:   params.Username,
			Email:      params.Username,
			Passwd:     params.Passwd,
			RegisterIp: c.ClientIP(),
			RegisterAt: time.Now(),
		},
		ReferInfo: models.ReferInfo{
			Parent:    params.ReferCode,
			ReferCode: nil, // don't support yet
		},
		RoleInfo: models.RoleInfo{
			Role: role,
		},
	}

	err = models.CreateUser(user)
	if err != nil {
		ResFailure(c, 500, err.Error())
		return
	}

	utils.Infof("[ Register ] success register one user: %v", params.Username)
	ResSuccess(c, nil)
}

func GetUserInfo(c *gin.Context) {
	user, err := token.GetUserInfo(c.Request) // 获取claims user
	if err != nil {
		ResFailure(c, 401, "Wrong Claims.")
		return
	}

	// 将信息从db中取出来更完整的信息返还
	info, err := models.GetUser(user.Name)
	if err != nil {
		ResFailure(c, 404, "The user is not exist.")
		return
	}

	// 如果没有 apikey, 则生成一个
	if info.ApiKey == "" {
		apiKey := utils.GenerateAsciiCode(32)
		err = models.UpdateUserApiKey(info.Username, apiKey)
		if err != nil {
			ResFailure(c, 401, "System error, please retry")
			return
		}
	}

	// 获取 login history
	histories, _ := models.GetUserLoginHistories(user.Name)

	data := struct {
		User         *models.User          `json:"user"`
		LoginHistory []models.LoginHistory `json:"loginHistory"`
	}{
		User:         info,
		LoginHistory: histories,
	}

	ResSuccess(c, data)
}

func ResetPassword(c *gin.Context) {
	user, err := token.GetUserInfo(c.Request) // 获取claims user
	if err != nil {
		ResFailure(c, 401, "Wrong Claims.")
		return
	}

	var rset models.ResetPasswdForm
	err = c.ShouldBind(&rset)
	if err != nil {
		ResFailure(c, 401, "Wrong Params.")
		return
	}

	// 转换密码为加密
	var err1 error
	rset.PasswdNew, err1 = utils.HashPassword(rset.PasswdNew)
	if err1 != nil {
		ResFailure(c, 401, "System error, please try again.")
		return
	}

	if !models.CheckUserPass(user.Name, rset.PasswdOld) {
		ResFailure(c, 401, "wrong old password")
		return
	}

	// updateDb
	err = models.ResetUserPassword(user.Name, rset.PasswdNew)
	if err != nil {
		ResFailure(c, 500, "System error to reset password")
		return
	}

	ResSuccess(c, "Reset password success")
}

func Test(c *gin.Context) {
	utils.Infof("Test pass, valid token.")

	ResSuccess(c, "Test Success.")
}
