package pair

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SavePairInfo(pair *schema.Pair, mongodb *mongo.Client) {
	log.Printf("[ SavePairInfo ] InsertOne now. pair: %v\n", pair.Address)

	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	pair.UpdatedAt = time.Now()
	pair.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), pair)
	if err != nil {
		// 冲突挺多，不打印了
		// log.Printf("[ SavePairInfo ] InsertOne error: %v, token: %v\n", err, pair.Address)
		return
	}
}

// 如果存在就更新, 不存在就插入
func UpSertPairInfo(pair *schema.Pair, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	pair.UpdatedAt = time.Now()

	filter := bson.D{{Key: "address", Value: pair.Address}}
	update := bson.D{{Key: "$set", Value: pair}}
	opt := options.Update().SetUpsert(true) // 执行更新操作，设置upsert为true

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		log.Printf("[ UpSertPairInfo ] failed. pair: %v, err: %v\n", pair.Address, err)
	} else {
		log.Printf("[ UpSertPairInfo ] success. pair: %v\n", pair.Address)
	}
}
