package models

import (
	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/mongo"

	"sfilter/user/config"
	"sfilter/utils"
)

var um UserModel

func InitService(db *mongo.Client) {
	if um.Collection == nil {
		um.Collection = db.Database(config.DatabaseName).Collection(config.UserDbName)
		initTable(um.Collection, config.UserDbName, UserIndexModel, db)
	}
}

func GetUser(id string) error {
	return nil
}

func CreateUser(user *token.User) error {
	return nil
}

// 用于 oauth 等场景, 直接获取或者直接注册
func GetOrCreateUser(user *token.User) (*User, error) {
	localUser := &User{}

	return localUser, nil
}

func CheckUserPass(username, passwd string) bool {
	user, err := um.GetUserByNameOrEmail(username)
	if err != nil {
		utils.Warnf("[ CheckUserPass ] GetUserByNameOrEmail err: %v", err)
		return false
	}

	if user.Password != passwd {
		utils.Warnf("[ CheckUserPass ] user(%v) password is not correct", username)
		return false
	}

	utils.Debugf("[ CheckUserPass ] user(%v) login success", username)
	return true
}
