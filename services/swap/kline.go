package swap

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func UpdateKline(swap *schema.Swap, mongodb *mongo.Client) {
}

func SaveKline(kline *schema.KLine5Min, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline5MinTableName)

	kline.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), kline)
	if err != nil {
		log.Printf("[ SaveKline ] InsertOne error: %v, kline: %v\n", err, kline)
		return
	} else {
		log.Printf("[ SaveKline ] InsertOne success, kline: %v\n", kline)
	}
}

func TEST_KLINE() {
	var pair = schema.KLinePairInfo{
		Symbol:     "pepe/weth",
		BaseToken:  "PEPE",
		QuoteToken: "WETH",
	}
	var k5min = schema.KLine5Min{
		KLinePairInfo: pair,
	}

	// SaveKline(&k5min, chain.GetMongo())

	filter := bson.M{"symbol": "pepe/weth"}
	collection := chain.GetMongo().Database(config.DatabaseName).Collection(config.Kline5MinTableName)

	err := collection.FindOne(context.Background(), filter).Decode(&k5min)

	log.Printf("err: %v, k5min: %v\n", err, k5min)
	log.Println("debug, symbol: ", k5min.Symbol)
}
