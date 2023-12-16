package controllers

import (
	"fmt"
	"sfilter/user/models"
	"time"
)

func preCheckRegister(form *models.RegisterForm) error {
	err := isValidCredentials(form)
	if err != nil {
		return err
	}

	// 检查验证码是否正确
	deadline := time.Now().Add(-10 * time.Minute) // 最近10min
	vcode, err := models.GetVerifyCodeByUser(form.Username, deadline)
	if err != nil || vcode.Code != form.AuthCode {
		return fmt.Errorf("wrong verify code")
	}

	// 目前beta阶段, 必须要有refercode 且refer人存在
	if form.ReferCode == "" || !models.IsExistedReferCode(form.ReferCode) {
		return fmt.Errorf("wrong params: refer code is not exist")
	}

	// username或email是否已存在
	if models.IsExistedUsernameOrEmail(form.Username) {
		return fmt.Errorf("username or email has existed, please change one")
	}

	return nil
}
