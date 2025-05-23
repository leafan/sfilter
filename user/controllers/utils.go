package controllers

import (
	"errors"
	"fmt"
	"regexp"
	"sfilter/config"
	"sfilter/user/models"
	"sfilter/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 预检查, 防止sql注入等
func isValidCredentials(form *models.RegisterForm) error {
	if len(form.Username) < 3 || len(form.Username) > 100 {
		return fmt.Errorf("wrong params: username too long or too short")
	}

	if containsSpecialCharacters(form.Username) {
		return fmt.Errorf("wrong params: username has special chars which deepeye not support")
	}

	if containsSpecialCharacters(form.ReferCode) {
		return fmt.Errorf("wrong params: refercode has special chars which deepeye not support")
	}

	if containsSpecialCharacters(form.AuthCode) {
		return fmt.Errorf("wrong params: authcode has special chars which deepeye not support")
	}

	// 检查是否是合法邮箱, 目前username就是email
	if !isValidEmail(form.Username) {
		return fmt.Errorf("wrong params: email address is illegal")
	}

	return nil
}

func containsSpecialCharacters(input string) bool {
	reg := regexp.MustCompile(`[^a-zA-Z0-9@.+-_%]`)
	return reg.MatchString(input)
}

func isValidEmail(email string) bool {
	regexPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(regexPattern, email)
	if err != nil {
		utils.Warnf("[ isValidEmail ] not valid, email: %v, err: %v", email, err)
		return false
	}
	return matched
}

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
	// if form.ReferCode == "" || !models.IsExistedReferCode(form.ReferCode) {
	// 	return fmt.Errorf("wrong params: refer code not exists")
	// }

	// username或email是否已存在
	if models.IsExistedUsernameOrEmail(form.Username) {
		return fmt.Errorf("username or email has existed, please change one")
	}

	return nil
}

func ParsePageLimitParams(c *gin.Context) (int64, int64, error) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("pageSize", "10")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return 0, 0, errors.New("invalid page parameter")
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return 0, 0, errors.New("invalid limit parameter")
	}

	if pageInt <= 0 {
		pageInt = 1
	}

	if pageInt > config.MONGO_PAGE_UPPER {
		pageInt = config.MONGO_PAGE_UPPER
	}

	if limitInt < config.MONGO_LIMIT_DOWN {
		limitInt = config.MONGO_LIMIT_DOWN
	}

	if limitInt > config.MONGO_LIMIT_UPPER {
		limitInt = config.MONGO_LIMIT_UPPER
	}

	return int64(pageInt), int64(limitInt), nil
}
