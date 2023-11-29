package global

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

const GlobalInfoKey = "global_info"

func GetGlobalInfo(mongodb *mongo.Client) (*schema.GlobalInfo, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.ConfigTableName)
	filter := bson.D{{Key: "configKey", Value: GlobalInfoKey}}

	var result schema.GlobalInfo
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	return &result, err
}

func UpdateGlobalInfo(info *schema.GlobalInfo, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.ConfigTableName)
	filter := bson.D{{Key: "configKey", Value: GlobalInfoKey}}
	opt := options.Update().SetUpsert(true)

	info.UpdatedAt = time.Now()

	update := bson.D{
		{Key: "$set", Value: info},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		log.Printf("[ UpsertTrends ] failed. err: %v\n", err)
	}

}
