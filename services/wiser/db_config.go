package wiser

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

const WiserConfigKey = "wiser_config"

func GetCurrentEpoch(mongodb *mongo.Database) (string, error) {
	cfg, err := GetWiserConfig(mongodb)
	if err != nil {
		return "", err
	}

	return cfg.Epoch, nil
}

// get用database, 因为可能有多个 db
func GetWiserConfig(mongodb *mongo.Database) (*schema.WiserDBConfig, error) {
	collection := mongodb.Collection(config.ConfigTableName)
	filter := bson.D{{Key: "configKey", Value: WiserConfigKey}}
	options := options.FindOne().SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	var result schema.WiserDBConfig
	err := collection.FindOne(context.Background(), filter, options).Decode(&result)

	return &result, err
}

func UpdateWiserConfig(info *schema.WiserDBConfig, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.ConfigTableName)
	filter := bson.D{{Key: "configKey", Value: WiserConfigKey}}
	opt := options.Update().SetUpsert(true)

	info.UpdatedAt = time.Now()

	update := bson.D{
		{Key: "$set", Value: info},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ UpdateGlobalInfo ] failed. err: %v\n", err)
	}

}
