package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Kline 基础结构体，只包含 高开低收
type KLineBasic struct {
	OpenPrice  float64 `bson:"openPrice" json:"openPrice"`
	ClosePrice float64 `bson:"closePrice" json:"closePrice"`
	HighPrice  float64 `bson:"highPrice" json:"highPrice"`
	LowPrice   float64 `bson:"lowPrice" json:"lowPrice"`
	Volume     float64 `bson:"volume" json:"volume"`
}

type KLinePairInfo struct {
	Symbol     string `bson:"symbol" json:"symbol"`         // 交易对, 如 PEPE/WETH, 大写
	BaseToken  string `bson:"baseToken" json:"baseToken"`   // base 币, 对应为 PEPE
	QuoteToken string `bson:"quoteToken" json:"quoteToken"` // quote计价币, 对应为 WETH
}

// KLine-5min 结构体, 为了减少数据行, 每个交易对保存一行, 多个点存一行
// 5min一个点, 一天最多288个点，每4小时一个字段(48个点)，一共 42个字段; 超过即删除
type KLine5Min struct {
	KLinePairInfo `bson:",inline"` // inline表示内连展开存储

	Kline1 KLineBasic

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

var KlineIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		// symbol 应该是唯一的
		Keys:    bson.D{{Key: "symbol", Value: 1}},
		Options: options.Index().SetName("symbol_index").SetUnique(true),
	},
	{
		// base token也就是main token, 需要查询. quoteToken就算了
		Keys:    bson.D{{Key: "baseToken", Value: 1}},
		Options: options.Index().SetName("baseToken_index").SetUnique(true),
	},
}
