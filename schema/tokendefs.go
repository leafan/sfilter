package schema

import (
	"math/big"
	"sfilter/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// basic token info
type Token struct {
	Address string
	Name    string
	Symbol  string

	TotalSupply *big.Int
	Decimal     uint8
}

var TokenIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdat", Value: -1}},
		Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(config.NeverExpireTime),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index").SetUnique(true),
	},
}
