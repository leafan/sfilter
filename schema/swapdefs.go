package schema

import (
	"context"
	"log"
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

const SwapSaveTime = 60 * 60 * 24 * 7 // 7d
// const SwapSaveTime = 10

var SwapIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{"createdat", -1}},
		Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(SwapSaveTime),
	},
	{
		Keys:    bson.D{{"tx", 1}},
		Options: options.Index().SetName("tx_index"),
	},
	{
		Keys:    bson.D{{"receiver", 1}},
		Options: options.Index().SetName("receiver_index"),
	},
}

// 如果不存在表, 则创建索引
func InitTables(mongodb *mongo.Client) {
	collection := mongodb.Database("deepeye").Collection("swap")

	filter := bson.M{"name": "your_collection_name"}
	_, err := collection.Database().ListCollectionNames(context.Background(), filter)
	if err != nil {
		log.Fatal("[ InitTables] ListCollectionNames err: ", err)
		return
	}

	if err == mongo.ErrNilDocument {
		// 说明是新表, 则创建索引
		_, err = collection.Indexes().CreateMany(context.Background(), SwapIndexModel)
		if err != nil {
			log.Printf("[ InitTables ] collection.Indexes().CreateMany error: %v\n", err)
			return
		}

		log.Printf("[ InitTables ] collection.Indexes().CreateMany success\n")
	} else {
		log.Printf("[ InitTables ] swap table exist, pass...")
	}
}
