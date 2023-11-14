package schema

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// v2 addr: https://etherscan.io/address/0x42d52847be255eacee8c3f96b3b223c0b3cc0438
// v3 addr: https://etherscan.io/address/0xea05d862e4c5cd0d3e660e0fcb2045c8dd4d7912

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
