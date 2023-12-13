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
	Password    string    `json:"-" bson:"password"`
	IsConfirmed bool      `json:"isConfirmed" bson:"isConfirmed"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

type ReferInfo struct {
	Parent    int    `json:"parent" bson:"parent"`
	ReferCode string `json:"referCode" bson:"referCode"`
	ReferNum  int    `json:"referNum" bson:"referNum"`
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
		Keys:    bson.D{{Key: "referCode", Value: -1}},
		Options: options.Index().SetName("referCode_index").SetSparse(true).SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "isConfirmed", Value: -1}},
		Options: options.Index().SetName("isConfirmed_index"),
	},
	{
		Keys:    bson.D{{Key: "parent", Value: -1}},
		Options: options.Index().SetName("parent_index"),
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
