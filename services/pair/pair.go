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

	pair.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), pair)
	if err != nil {
		// 冲突挺多，不打印了
		// log.Printf("[ SavePairInfo ] InsertOne error: %v, token: %v\n", err, pair.Address)
		return
	}
}

/**
 * reference:
update := bson.D{
        {Key: "$set", Value: bson.D{
            {Key: "field1", Value: pair.Field1}, //更新部分字段
            {Key: "field2", Value: pair.Field2},
            // 添加其他需要更新的字段
        }},
        {Key: "$setOnInsert", Value: pair}, // 如果文档不存在，则插入整个数据
    }
*/

// 如果存在就更新, 不存在就插入
func UpSertPairInfo(pair *schema.Pair, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

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
