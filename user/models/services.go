package models

import (
	"time"

	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/mongo"

	"sfilter/user/config"
	"sfilter/utils"
)

var um UserModel
var vm VerifyCodeModel
var lm LoginHistoryModel

func InitService(db *mongo.Client) {
	// 初始化table
	initTables(db)

	// 初始化service
	if um.Collection == nil {
		um.Collection = db.Database(config.DatabaseName).Collection(config.UserTableName)
	}

	if vm.Collection == nil {
		vm.Collection = db.Database(config.DatabaseName).Collection(config.VerifyCodeTableName)
	}

	if lm.Collection == nil {
		lm.Collection = db.Database(config.DatabaseName).Collection(config.LoginHistoryTableName)
	}

	// 初始化数据
	checkOrCreatAdmin()
}

func GetUser(username string) (*User, error) {
	user, err := um.GetUserByNameOrEmail(username)
	if err != nil {
		utils.Warnf("[ GetUser ] GetUser err: %v", err)
		return nil, err
	}

	return user, nil
}

func CreateUser(user *User) error {
	err := um.CreatUser(user)
	return err
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

	if user.Passwd != passwd {
		utils.Warnf("[ CheckUserPass ] user(%v) password is not correct", username)
		return false
	}

	utils.Debugf("[ CheckUserPass ] user(%v) login success", username)
	return true
}

func IsExistedReferCode(refercode string) bool {
	if refercode == "" {
		return true // 不允许空值
	}

	_, err := um.GetUserByReferCode(refercode)
	return err == nil
}

func IsExistedUsernameOrEmail(username string) bool {
	_, err := um.GetUserByNameOrEmail(username)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			utils.Errorf("[ IsExistedUsernameOrEmail ] conn db error: %v", err)
			return true
		} else {
			return false
		}
	}

	// utils.Tracef("[ IsExistedUsernameOrEmail ] exist user: %v, err: %v", user, err)
	return true
}

func ResetUserPassword(username, newpass string) error {
	return um.ResetUserPassword(username, newpass)
}

// for verifycode, 只获取最近10分钟的有效数据
func GetVerifyCodeByUser(username string, deadline time.Time) (*VerifyCode, error) {
	return vm.GetCodeByUsername(username, deadline)
}

// 针对ip获取验证码历史
func GetVerifyCodesByip(ip string, deadline time.Time) ([]VerifyCode, error) {
	return vm.GetCodesByIp(ip, deadline)
}

func CreatVerifyCode(username, code, ip string) error {
	vcode := &VerifyCode{
		Username: username,
		Code:     code,
		ClientIp: ip,
	}

	return vm.CreatCode(vcode)
}

func GetUserLoginHistories(username string) ([]LoginHistory, error) {
	return lm.GetHistoriesByUsername(username)
}

func CreatOneLoginHistory(username, ip string) error {
	entry := LoginHistory{
		Username: username,
		LoginIp:  ip,
	}

	return lm.CreatOne(&entry)
}
