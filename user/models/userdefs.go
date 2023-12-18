package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	BasicInfo `bson:",inline"`
	ReferInfo `bson:",inline"`
	RoleInfo  `bson:",inline"`
}

type BasicInfo struct {
	Username    string    `json:"username" bson:"username"`
	Nickname    string    `json:"nickname" bson:"nickname"`
	Email       string    `json:"email" bson:"email"`
	Address     string    `json:"address" bson:"address"`
	Phone       string    `json:"phone" bson:"phone"`
	Passwd      string    `json:"-" bson:"passwd"`
	RegisterIp  string    `json:"-" bson:"registerIp"`
	IsConfirmed bool      `json:"-" bson:"isConfirmed"`
	RegisterAt  time.Time `json:"registerAt" bson:"registerAt"`
	UpdatedAt   time.Time `json:"-" bson:"updatedAt"`
}

type ReferInfo struct {
	Parent    string  `json:"-" bson:"parent"`            // 我的邀请人
	ReferCode *string `json:"referCode" bson:"referCode"` // 我的邀请码
	ReferNum  int     `json:"-" bson:"referNum"`          // 我的邀请人数, 暂时不用
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
			SetPartialFilterExpression(bson.M{"referCode": bson.M{"$type": "string"}}),
	},
	{
		Keys:    bson.D{{Key: "registerAt", Value: -1}},
		Options: options.Index().SetName("registerAt_index"),
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
