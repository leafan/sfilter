package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Pair struct {
	Address string `json:"address" bson:"address"` // 地址
	Token0  string `json:"token0" bson:"token0"`   // 代币0
	Token1  string `json:"token1" bson:"token1"`   // 代币1

	Type int `json:"type" bson:"type"` // 类型是uniswap v2还是v3等

	PairCreatedBlockNo uint64 `json:"pairCreatedBlockNo" bson:"pairCreatedBlockNo"`
	PairCreatedTime    int64  `json:"pairCreatedTime" bson:"pairCreatedTime"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"` // 创建时间
}

var PairIndexModel = []mongo.IndexModel{
	{
		Keys: bson.D{{Key: "createdAt", Value: -1}},
		// Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(config.NeverExpireTime),
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index").SetUnique(true),
	},
}
