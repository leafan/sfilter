package schema

import (
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Swap struct {
	Tx        string // 原始hash
	Token     string // 买卖的token
	Direction int    // 0买; 1卖

	Operator string // msg.sender0
	Receiver string //接收token的人，一般和operator相等, 但也可能为合约

	Price         *big.Int
	AmountInToken *big.Int
	CreatedAt     time.Time
}

const swapSaveTime = 60 * 60 * 24 * 7 // 7d
// const swapSaveTime = 10

var SwapIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdat", Value: -1}},
		Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(swapSaveTime),
	},
	{
		Keys:    bson.D{{Key: "tx", Value: 1}},
		Options: options.Index().SetName("tx_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "receiver", Value: 1}},
		Options: options.Index().SetName("receiver_index"),
	},
}
