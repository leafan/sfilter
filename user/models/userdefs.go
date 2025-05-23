package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 用户状态
const (
	USER_STATUS_NORMAL int = iota // 正常

	// 跟踪功能被禁用, 比如跟踪数量太大了
	USER_STATUS_TRACK_BLOCKED

	// ...

	USER_STATUS_UNKNOWN = 9999 // root管理员用户
)

type User struct {
	BasicInfo `json:",inline" bson:",inline"`
	ReferInfo `json:",inline" bson:",inline"`
	RoleInfo  `json:",inline" bson:",inline"`
}

type BasicInfo struct {
	Username       string `json:"username" bson:"username"`
	Nickname       string `json:"nickname" bson:"nickname"`
	Email          string `json:"email" bson:"email"`
	Address        string `json:"address" bson:"address"`
	Phone          string `json:"phone" bson:"phone"`
	Passwd         string `json:"-" bson:"passwd"`
	RegisterIp     string `json:"registerIp" bson:"registerIp"`
	RegisterRegion string `json:"registerRegion" bson:"registerRegion"`
	IsConfirmed    bool   `json:"-" bson:"isConfirmed"` // reserved

	// 该字段表示用户状态, 比如是否被禁用等
	Status     int       `json:"status" bson:"status"`
	RegisterAt time.Time `json:"registerAt" bson:"registerAt"`
	UpdatedAt  time.Time `json:"-" bson:"updatedAt"`
}

type ReferInfo struct {
	Parent    string  `json:"parent" bson:"parent"` // 我的邀请人
	ReferCode *string `json:"-" bson:"referCode"`   // 我的邀请码
	ReferNum  int     `json:"-" bson:"referNum"`    // 我的邀请人数, 暂时不用
}

type RoleInfo struct {
	ApiKey    string    `json:"apiKey" bson:"apiKey"`
	Role      int       `json:"role" bson:"role"`
	ExpiredAt time.Time `json:"expiredAt" bson:"expiredAt"`
}

var UserIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "username", Value: -1}},
		Options: options.Index().SetName("username_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "email", Value: -1}},
		Options: options.Index().SetName("email_index").SetSparse(true).SetUnique(true),
	},
	{
		Keys: bson.D{{Key: "parent", Value: -1}},
		Options: options.Index().SetName("parent_index").
			SetSparse(true),
	},
	{
		Keys: bson.D{{Key: "referCode", Value: -1}},
		Options: options.Index().SetName("referCode_index").
			SetUnique(true).
			SetPartialFilterExpression(bson.M{"apiKey": bson.M{"$gt": ""}}),
	},
	{
		Keys:    bson.D{{Key: "registerAt", Value: -1}},
		Options: options.Index().SetName("registerAt_index"),
	},
	{
		Keys: bson.D{{Key: "apiKey", Value: -1}},
		Options: options.Index().SetName("apiKey_index").SetUnique(true).
			SetPartialFilterExpression(bson.M{"apiKey": bson.M{"$gt": ""}}),
	},
	{
		Keys:    bson.D{{Key: "referNum", Value: -1}},
		Options: options.Index().SetName("referNum_index"),
	},
	{
		Keys:    bson.D{{Key: "role", Value: -1}},
		Options: options.Index().SetName("role_index"),
	},
	{
		Keys:    bson.D{{Key: "expiredAt", Value: -1}},
		Options: options.Index().SetName("expiredAt_index"),
	},
}
