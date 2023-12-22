package auth

import (
	"fmt"
	"sfilter/utils"

	"sfilter/user/models"

	"github.com/go-pkgz/auth/token"
)

// 暂时不用, 如黑名单等
// middleware 中判断token成功后使用
func ValidateTokenClaims(t string, claims token.Claims) bool {
	// utils.Warnf("[ ValidateTokenClaims ] trace here. user: %v", claims.User.Name)

	return true
}

// 登陆成功时, 更新用户的一些数据到token
// 比如他的级别(决定可以访问哪些内容等)
func UserClaimsUpdate(userDataFetcher func(user *token.User) (*models.User, error)) token.ClaimsUpdFunc {
	return func(claims token.Claims) token.Claims {
		// utils.Tracef("[ UserClaimsUpdate ] Debug. claims here: %v", claims)

		if claims.User == nil {
			return claims
		}

		userData, err := userDataFetcher(claims.User)
		if err != nil {
			return claims
		}

		claims.User.SetStrAttr("role", fmt.Sprintf("%d", userData.Role))

		utils.Tracef("[ UserClaimsUpdate ] login success ip: %v, user: %v", claims.User.IP, claims.User)

		return claims
	}
}
