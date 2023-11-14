package schema

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Pair struct {
	Address string
	Token0  string
	Token1  string

	Type int // 是uniswap v2还是v3 etc
}

var PairIndexModel = []mongo.IndexModel{
	{
		Keys: bson.D{{Key: "createdat", Value: -1}},
		// Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(config.NeverExpireTime),
		Options: options.Index().SetName("createdat_index"),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index").SetUnique(true),
	},
}
