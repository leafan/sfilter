package services

/**
 * @desc 用 aws 的 ses 系统发送邮件验证码
 */

import (
	"net/http"
	"sfilter/user/config"
	"sfilter/utils"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const CharSet = "UTF-8"

type AwsService struct {
	Client *ses.SES
}

func NewAwsClient() *ses.SES {
	awsConf := &aws.Config{}

	awsConf.Region = aws.String(config.AWS_REGION)
	awsConf.MaxRetries = aws.Int(1)
	awsConf.EnableEndpointDiscovery = aws.Bool(true)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	awsConf.HTTPClient = httpClient // 将自定义的 http.Client 设置到 aws.Config 中

	awsConf.Credentials = credentials.NewStaticCredentials(config.AWS_KEY_ID, config.AWS_SECRET_KEY, "")

	awsSess := session.Must(session.NewSession(awsConf))

	client := ses.New(awsSess, awsConf)

	return client
}

func (s *AwsService) SendMail(email *Email) error {
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: aws.StringSlice(email.Cc),
			ToAddresses: aws.StringSlice(email.To),
		},

		// html和text都设置时, 由客户端决定展示哪一种
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(email.HTML),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(email.Text),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(email.Subject),
			},
		},
		Source: aws.String(email.From),
	}

	_, err := s.Client.SendEmail(input)

	if err != nil {
		utils.Errorf("[ SendMail ] send error: %v, email: %v", err, email)
		return err
	}

	utils.Infof("[ SendMail ]  sent to address: %v success", email.To)
	return nil
}
