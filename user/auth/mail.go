package auth

/**
 * @desc 用 https://app.mailersend.com 系统发送邮件验证码
 */

import (
	"context"
	"sfilter/user/config"
	"sfilter/utils"
	"time"

	"github.com/mailersend/mailersend-go"
)

var mailerInstance *mailersend.Mailersend = nil

func getMailer() *mailersend.Mailersend {
	if mailerInstance == nil {
		mailerInstance = NewMailer()
	}

	return mailerInstance
}

func NewMailer() *mailersend.Mailersend {
	ms := mailersend.NewMailersend(config.MailerApiKey)
	if ms == nil {
		utils.Fatalf("[ NewMailer ] NewMailersend failed.. key: ", config.MailerApiKey)
	}

	return ms
}

func SendMailByTemplate(to, templateId string) error {
	ms := getMailer()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	subject := "Register Code"

	from := mailersend.From{
		Name:  "noreply",
		Email: "noreply.deepeye.cc",
	}

	recipients := []mailersend.Recipient{
		{
			Name:  "Client",
			Email: to,
		},
	}

	variables := []mailersend.Variables{
		{
			Email: to,
			Substitutions: []mailersend.Substitution{
				{
					Var:   "authenticate_code",
					Value: "login_code",
				},
			},
		},
	}

	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetTemplateID(templateId)
	message.SetSubstitutions(variables)

	ret, err := ms.Email.Send(ctx, message)
	if err != nil {
		utils.Errorf("[ SendMailByTemplate ] send error: %v", err)
		return err
	}

	utils.Infof("[ SendMailByTemplate ] send success. ret: %v", ret)
	return nil
}

func TEST_EMAIL() {
	to := "market@deepeye.cc"
	tid := "v69oxl5zepdl785k"

	SendMailByTemplate(to, tid)
}
