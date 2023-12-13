package auth

import (
	"errors"
	"regexp"
	"sfilter/user/models"
	"sfilter/utils"
	"strings"

	"github.com/go-pkgz/auth/provider"
)

func PasswordCredCheck() provider.CredCheckerFunc {
	var isValidAnonName = regexp.MustCompile(`^[a-zA-Z][\w ]+$`).MatchString

	return func(user, passwd string) (ok bool, err error) {
		utils.Debugf("[ PasswordCredCheck ] user: %v, passwd: %v", user, passwd)

		user = strings.TrimSpace(user)
		if len(user) < 3 {
			utils.Warnf("name %q is too short, should be at least 3 characters", user)
			return false, errors.New("name is invalid")
		}

		if !isValidAnonName(user) {
			utils.Warnf("[WARN] name %q should have letters, digits, underscores and spaces only", user)
			return false, errors.New("name is invalid")
		}

		// 检查密码和db中是否一致
		isUser := models.CheckUserPass(user, passwd)

		return isUser, nil
	}
}
