package auth

/**
 * 通过如下服务生成claims, 同时会设置cookie, 将jwt设置到cookie中, 字段为 JWT

 * postman 需要请求测试时, 在 header 中带上 X-JWT
 * 前端自动cookie中获取请求即可, cookie使用时, 要同时带上 XSRF-TOKEN

 */

import (
	"fmt"
	"sfilter/user/config"
	"sfilter/user/models"
	"sfilter/utils"
	"strings"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/logger"
	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewAuthService(db *mongo.Client) *auth.Service {
	models.InitService(db)

	opts := getAuthOptions()
	opts.ClaimsUpd = UserClaimsUpdate(models.GetOrCreateUser)

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
	}
}

// 暂时不用
func ValidateTokenClaims(t string, claims token.Claims) bool {
	utils.Warnf("[ ValidateTokenClaims ] test here. user: %v", claims.User)

	if claims.User != nil {
		// 如黑名单处理, 临时配置下发等
		return strings.HasPrefix(claims.User.Name, "test")
	}

	return false
}

// 更新用户的一些数据到token
// 比如他的级别(决定可以访问哪些内容等)
func UserClaimsUpdate(userDataFetcher func(user *token.User) (*models.User, error)) token.ClaimsUpdFunc {
	return func(claims token.Claims) token.Claims {
		utils.Tracef("[ UserClaimsUpdate ] Debug. claims here: %v", claims)

		if claims.User == nil {
			return claims
		}

		userData, err := userDataFetcher(claims.User)
		if err != nil {
			return claims
		}

		claims.User.SetStrAttr("role", fmt.Sprintf("%d", userData.Role))

		return claims
	}
}
