package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// basic token info
type Token struct {
	Address     string `json:"address" bson:"address"`         // 地址
	Name        string `json:"name" bson:"name"`               // 名称
	Symbol      string `json:"symbol" bson:"symbol"`           // 符号
	TotalSupply string `json:"totalSupply" bson:"totalSupply"` // 总供应量
	Decimal     uint8  `json:"decimal" bson:"decimal"`         // 小数位数
	FDV         string `json:"fdv" bson:"fdv"`                 // 流通市值

	// 直接float存储, 毕竟这里只是展示, 不需要高精度计算
	// K线能计算就更新, 否则为0(比如pair为2个屌丝币)
	PriceInUsd float64 `json:"priceInUsd" bson:"priceInUsd"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"` // 更新时间
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"` // 创建时间
}

var TokenIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
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
