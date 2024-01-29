package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 定义Token map, 避免反复查db
type TokenMap map[string]*Token

type TokenInfoOnChain struct {
	Address string `json:"address" bson:"address"` // 地址
	Name    string `json:"name" bson:"name"`       // 名称
	Symbol  string `json:"symbol" bson:"symbol"`   // 符号

	PriceInUsd float64 `json:"priceInUsd" bson:"priceInUsd"`

	TotalSupply string `json:"totalSupply" bson:"totalSupply"` // 总供应量
	Decimal     uint8  `json:"decimal" bson:"decimal"`         // 小数位数
}

// 第三方数据获
type TokenInfoOn3rdParty struct {
	FDV string `json:"fdv" bson:"fdv"` // 流通市值

	OfficialSite string `json:"officialSite" bson:"officialSite"`

	// etc
}

// basic token info
type Token struct {
	TokenInfoOnChain `bson:",inline"`

	TokenInfoOn3rdParty `bson:",inline"`

	UpdatedAt time.Time `json:"-" bson:"updatedAt"` // 更新时间
	CreatedAt time.Time `json:"-" bson:"createdAt"` // 创建时间
}

var TokenIndexModel = []mongo.IndexModel{
	{
		// 以updateAt为键值, 目的是超过一定时间无更新的作为废弃Token删除
		// 防止表过大
		Keys:    bson.D{{Key: "UpdatedAt", Value: -1}},
		Options: options.Index().SetName("UpdatedAt_index").SetExpireAfterSeconds(config.TokenTableSavetime),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetName("name_index"),
	},
	{
		Keys:    bson.D{{Key: "symbol", Value: 1}},
		Options: options.Index().SetName("symbol_index"),
	},
}
