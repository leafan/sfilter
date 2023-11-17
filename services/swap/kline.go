package swap

import (
	"context"
	"fmt"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateKline(swap *schema.Swap, mongodb *mongo.Client) {
	update5MinKline(swap, mongodb)
	update1DayKline(swap, mongodb)
}

func update5MinKline(swap *schema.Swap, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline5MinTableName)

	symbol := fmt.Sprintf("%v/%v", swap.Token0, swap.Token1)
	if swap.MainToken == swap.Token1 {
		symbol = fmt.Sprintf("%v/%v", swap.Token1, swap.Token0)
	}
	key := fmt.Sprintf("%v_%v", symbol, time.Unix(swap.SwapTime, 0).Day())
	fmt.Printf("[ update5MinKline ] key: %v\n", key)

	filter := bson.M{"symbolDay": key}

	var kline schema.KLines5Min
	err := collection.FindOne(context.Background(), filter).Decode(&kline)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("[ update5MinKline ] FindOne error: %v, swap tx: %v\n", err, swap.LogIndexWithTx)
		return
	}

	// 更新具体数据并写入
	hour := time.Unix(swap.SwapTime, 0).Hour()
	data := schema.Get5MinFieldByHour(&kline, hour)

	newKline := schema.KLine{
		OpenPrice:  1.1,
		ClosePrice: 2.2,
		Volume:     5.5,
	}

	data = append(data, &newKline)
	fmt.Printf("[ update5MinKline ] after Hour: %v, kline: %v\n", hour, kline)

	SaveUpsert5MinKline(&kline, mongodb)
}

func update1DayKline(swap *schema.Swap, mongodb *mongo.Client) {

}

func SaveUpsert5MinKline(kline *schema.KLines5Min, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline5MinTableName)

	kline.UpdatedAt = time.Now()

	filter := bson.D{{Key: "symbolDay", Value: kline.SymbolDay}}
	update := bson.D{{Key: "$set", Value: kline}}
	opt := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		log.Printf("[ SaveUpsert5MinKline ] InsertOne error: %v, kline: %v\n", err, kline)
		return
	} else {
		log.Printf("[ SaveUpsert5MinKline ] InsertOne success, kline: %v\n", kline)
	}
}

func TEST_KLINE() {
	var pair = schema.KLinePairInfo{
		Symbol:     "pepe/weth",
		BaseToken:  "PEPE",
		QuoteToken: "WETH",
	}
	var k5min = schema.KLines5Min{
		KLinePairInfo: pair,
	}

	// SaveKline(&k5min, chain.GetMongo())

	filter := bson.M{"symbol": "pepe/weth"}
	collection := chain.GetMongo().Database(config.DatabaseName).Collection(config.Kline5MinTableName)

	err := collection.FindOne(context.Background(), filter).Decode(&k5min)

	log.Printf("err: %v, k5min: %v\n", err, k5min)
	log.Println("debug, symbol: ", k5min.Symbol)
}
