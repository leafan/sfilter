package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RouterMap map[string]Router

type Router struct {
	Address   string    `json:"address" bson:"address"`
	Factory   string    `json:"factory" bson:"factory"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

var RouterIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index"),
	},
	{
		Keys:    bson.D{{Key: "factory", Value: 1}},
		Options: options.Index().SetName("factory_index"),
	},
}

type SpecialAddressMap map[string]SpecialAddress

const (
	SPECIAL_ADDRESS_TYPE_UNINIT int = iota

	SPECIAL_ADDRESS_TYPE_BLACK_HOLE     // 黑洞地址
	SPECIAL_ADDRESS_TYPE_DEFLAT_RECEIVE // 通缩币收币地址

)

type SpecialAddress struct {
	Address string `json:"address" bson:"address"`
	Desc    string `json:"desc" bson:"desc"` // 如factory等信息可以写到里面
	Type    int    `json:"type" bson:"type"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

var SpecialAddressIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index"),
	},
	{
		Keys:    bson.D{{Key: "type", Value: 1}},
		Options: options.Index().SetName("type_index"),
	},
}
