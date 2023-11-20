package kline

import (
	"context"
	"fmt"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 获取k线, 且倒推几小时且返回值必须生成几根柱子
func Get1MinKlineByPairForHours(pair string, end time.Time, hours int, mongodb *mongo.Client) []schema.KLinesForHour {
	start := end.Add(-time.Duration(hours) * time.Hour)
	data := get1MinKlineByPair(pair, start, end, mongodb)

	// 需要生成(hours+1)个柱子, 按时间正序
	// 如 1:10~3:10, 则有1,2,3点钟三根柱子, 即使是 1:00~3:00, 亦是如此
	var klines []schema.KLinesForHour
	for i := 0; i < hours+1; i++ {
		var kline schema.KLinesForHour

		// 判断当前时间柱子是否有查出数值
		klineTime := start.Add(time.Duration(i) * time.Hour)
		pairDayHour := fmt.Sprintf("%v_%v_%v", pair, klineTime.Day(), klineTime.Hour())

		for _, v := range data {
			// 如果当前小时有合法柱子, 拷贝, 否则全为0
			if v.PairDayHour == pairDayHour {
				// log.Printf("[ Get1MinKlineByPairForHours ] find one pairDayHour: %v, v: %v\n", pairDayHour, v)

				utils.DeepCopy(&v.Kline, &kline)
				break
			}
		}

		klines = append(klines, kline)
	}

	return klines
}

// 取出从start到end的所有柱子
func get1MinKlineByPair(pair string, start, end time.Time, mongodb *mongo.Client) []schema.KLines1Min {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1MinTableName)

	filter := bson.M{
		"pair": pair,
		"timestamp": bson.M{
			"$gte": start, "$lt": end,
		},
	}
	options := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	// log.Printf("[ get1MinKlineByPair ] debug filter: %v\n", filter)

	var result []schema.KLines1Min
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("[ get1MinKlineByPair ] Find error: %v, filter: %v\n", err, filter)
		}

		return result
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &result); err != nil {
		log.Printf("[ get1MinKlineByPair ] cursor.All error: %v, filter: %v\n", err, filter)
	}
	// log.Printf("[ get1MinKlineByPair ] result: %v\n\n", result)

	return result
}

func SaveUpsert1MinKline(kline *schema.KLines1Min, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1MinTableName)

	kline.UpdatedAt = time.Now()

	filter := bson.D{{Key: "pairDayHour", Value: kline.PairDayHour}}
	update := bson.D{{Key: "$set", Value: kline}}
	opt := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		log.Printf("[ SaveUpsert1MinKline ] InsertOne error: %v, kline: %v\n", err, kline)
		return
	}
}

func TEST_KLINE_DB() {
	pair := "0x00D05F6176aFa7c1CFe45a5629e1Eb0F6A5519b0"
	data := Get1MinKlineByPairForHours(pair, time.Now(), 5, chain.GetMongo())
	fmt.Println("Get1MinKlineByPairForHours: ", (data))
}
