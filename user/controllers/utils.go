package controllers

import (
	"fmt"
	"regexp"
	"sfilter/user/models"
	"sfilter/utils"
)

// 预检查, 防止sql注入等
func isValidCredentials(form *models.RegisterForm) error {
	if len(form.Username) < 3 || len(form.Username) > 30 {
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
	reg := regexp.MustCompile(`[^a-zA-Z0-9@.-]`)
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
