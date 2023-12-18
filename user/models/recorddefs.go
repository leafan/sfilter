package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VerifyCode struct {
	Username       string `json:"username" bson:"username"`
	Code           string `json:"code" bson:"code"`
	ClientIp       string `json:"clientIp" bson:"clientIp"`
	ClientLocation string `json:"clientLocation" bson:"clientLocation"`

	Status    int       `json:"status" bson:"status"` // 状态，目前未使用
	CreatedAt time.Time `json:"-" bson:"createdAt"`   // 创建时间
}

var VerifyCodeIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "clientIp", Value: -1}},
		Options: options.Index().SetName("clientIp_index"),
	},
	{
		Keys:    bson.D{{Key: "username", Value: -1}},
		Options: options.Index().SetName("username_index"),
	},
}

// 登陆历史记录
type LoginHistory struct {
	Username      string `json:"username" bson:"username"`
	LoginIp       string `json:"loginIp" bson:"loginIp"`
	LoginLocation string `json:"loginLocation" bson:"loginLocation"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"` // 创建时间
}

var LoginHistoryIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "username", Value: -1}},
		Options: options.Index().SetName("username_index"),
	},
}
