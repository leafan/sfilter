package models

import (
	"errors"
	"time"

	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"sfilter/user/config"
	"sfilter/user/services"
	"sfilter/utils"
)

var um UserModel
var vm VerifyCodeModel
var lm LoginHistoryModel
var tm TrackAddressModel

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

	if tm.Collection == nil {
		tm.Collection = db.Database(config.DatabaseName).Collection(config.TrackAddressTableName)
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

func GetUserInfoByAPIKey(apiKey string) (*User, error) {
	if apiKey == "" {
		return nil, errors.New("wrong apikey")
	}

	user, err := um.GetUserByApiKey(apiKey)
	if err != nil {
		utils.Warnf("[ GetUserInfoByAPIKey ] GetUser error. apikey: %v, err: %v", apiKey, err)
		return nil, err
	}

	return user, nil
}

func CreateUser(user *User) error {
	var err error
	user.RegisterRegion, err = services.GetIpLocation(user.RegisterIp)
	if err != nil {
		utils.Errorf("[ CreateUser ] getip location failed, err: %v", err)
		return err
	}

	err = um.CreatUser(user)
	return err
}

// 用于 oauth 等场景, 直接获取或者直接注册
func GetUserByToken(user *token.User) (*User, error) {
	dbUser, err := um.GetUserByNameOrEmail(user.Name)
	if err != nil {
		utils.Warnf("[ GetUserByToken ] GetUser err: %v", err)
		return nil, err
	}

	return dbUser, nil
}

func CheckUserPass(username, passwd string) bool {
	user, err := um.GetUserByNameOrEmail(username)
	if err != nil {
		utils.Warnf("[ CheckUserPass ] GetUserByNameOrEmail err: %v", err)
		return false
	}

	err = utils.ComparePassword(user.Passwd, passwd)
	if err != nil {
		utils.Warnf("[ CheckUserPass ] user(%v) password is not correct", username)
		return false
	}

	utils.Debugf("[ CheckUserPass ] user(%v) check success", username)
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

func UpdateUserApiKey(username, apikey string) error {
	return um.UpdateUserApiKey(username, apikey)
}

func ResetUserRole(username string, role int) error {
	return um.ResetUserRole(username, role)
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

	var err error
	vcode.ClientLocation, err = services.GetIpLocation(vcode.ClientIp)
	if err != nil {
		utils.Errorf("[ CreatVerifyCode ] getip location failed, err: %v", err)
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

	var err error
	entry.LoginLocation, err = services.GetIpLocation(entry.LoginIp)
	if err != nil {
		utils.Errorf("[ CreatOneLoginHistory ] getip location failed, err: %v", err)
	}

	return lm.CreatOne(&entry)
}

func GetTrackedAddresses(username string) ([]UserTrackedAddress, int64, error) {
	return tm.GetTrackAddressesByUsername(username, nil, nil)
}

func GetTrackedAddressesWithOptionFilter(username string, options *options.FindOptions, filter *primitive.M) ([]UserTrackedAddress, int64, error) {
	return tm.GetTrackAddressesByUsername(username, options, filter)
}

func CreateTrackedAddress(username, address, memo string, priority int) error {
	taddr := &UserTrackedAddress{
		Username: username,
		AddressInfo: AddressInfo{
			Address:  address,
			Memo:     memo,
			Priority: priority,
		},
	}

	return tm.CreatOne(taddr)
}

func GetEntryByUserAndMemo(username, memo string) (*UserTrackedAddress, error) {
	return tm.GetEntryByUserAndMemo(username, memo)
}

func GetEntryByUserAndAddress(username, address string) (*UserTrackedAddress, error) {
	return tm.GetEntryByUserAndAddress(username, address)
}

func GetUserTrackAddressCount(username string) (int64, error) {
	return tm.GetUserTrackAddressCount(username)
}

func UpdateTrackedAddress(username, address, memo string, priority int) error {
	addrinfo := &AddressInfo{
		Memo:     memo,
		Priority: priority,
	}

	return tm.UpdateTrackedAddress(username, address, addrinfo)
}

func DeleteTrackedAddress(username, address string) error {
	return tm.DeleteByUsernameAndAddr(username, address)
}

// //// Admin /////
func AdminGetAllUsersWithOptionFilter(options *options.FindOptions, filter *primitive.M) ([]User, int64, error) {
	return um.GetAllUsers(options, filter)
}

func AdminGetTrackAddressMap() (UserTrackAddressMap, int, error) {
	return tm.GetTrackAddressMap(1, 1000)
}

func AdminGetAllTrackAddrsWithOptionFilter(options *options.FindOptions, filter *primitive.M) ([]UserTrackedAddress, int64, error) {
	return tm.GetAllTrackAddrs(options, filter)
}
