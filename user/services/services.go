package services

import (
	"fmt"
	"sfilter/utils"
)

var service = AwsService{
	Client: NewAwsClient(),
}

func SendVerifyEmail(to, code string) error {
	var email Email
	utils.DeepCopy(&RegisterCodeEmailTemplate, &email)

	email.To = append(email.To, to)
	email.Text = fmt.Sprintf(email.Text, code)
	email.HTML = fmt.Sprintf(email.HTML, code)

	return service.SendMail(&email)
}

func TEST_EMAIL_AWS() {
	to := "leafan@qq.com"
	code := "515637"

	SendVerifyEmail(to, code)
}
