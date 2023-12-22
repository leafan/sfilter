package auth

/**
 * 通过如下服务生成claims, 同时会设置cookie, 将jwt设置到cookie中, 字段为 JWT

 * postman 需要请求测试时, 在 header 中带上 X-JWT
 * 前端自动cookie中获取请求即可, cookie使用时, 要同时带上 XSRF-TOKEN

 */

import (
	"sfilter/user/config"
	"sfilter/user/models"

	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/logger"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewAuthService(db *mongo.Client) *auth.Service {
	models.InitService(db)

	opts := getAuthOptions()
	opts.ClaimsUpd = UserClaimsUpdate(models.GetUserByToken)

	// validator allows to reject some valid tokens with user-defined logic
	opts.Validator = token.ValidatorFunc(ValidateTokenClaims)

	return auth.NewService(opts)
}

func getAuthOptions() auth.Opts {
	return auth.Opts{
		SecretReader: token.SecretFunc(func(id string) (string, error) {
			return config.JWTSecret, nil
		}),

		TokenDuration:  config.TokenDuration,
		CookieDuration: config.CookieDuration,
		Issuer:         config.Issuer,
		URL:            config.APIURL,
		AvatarStore:    avatar.NewNoOp(),
		SecureCookies:  true,
		AudSecrets:     false,
		Logger:         logger.Std,

		// 改为header传输key, cookie不会
		SendJWTHeader: true,
		JWTHeaderKey:  config.JWTHeaderKey,
	}
}
