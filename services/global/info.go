package global

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const GlobalInfoKey = "global_info"

// get用database, 因为可能有多个 db
func GetGlobalInfo(mongodb *mongo.Database) (*schema.GlobalInfo, error) {
	collection := mongodb.Collection(config.ConfigTableName)
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
		utils.Warnf("[ UpsertTrends ] failed. err: %v\n", err)
	}

}
