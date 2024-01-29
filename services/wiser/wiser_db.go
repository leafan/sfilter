package wiser

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveWiser(wiser *schema.Wiser, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.WiserTableName)

	wiser.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), wiser)

	if err != nil {
		// 如果是重复键值导致的插入失败
		filter := bson.D{{Key: "addressWithEpoch", Value: wiser.AddressWithEpoch}}
		opts := options.Update().SetUpsert(true)

		update := bson.D{
			{Key: "$set", Value: wiser},
		}
		_, err := collection.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			utils.Errorf("[ SaveWiser ] failed. address: %v, err: %v\n", wiser.
				Address, err)
		}
	}
}

func ResetWiserEpochData(mongodb *mongo.Client, epoch string) error {
	collection := mongodb.Database(config.DatabaseName).Collection(config.WiserTableName)
	deleteFilter := bson.D{{Key: "epoch", Value: epoch}}
	_, err := collection.DeleteMany(context.Background(), deleteFilter)
	if err != nil {
		utils.Errorf("[ DeleteOldEntries ] DeleteMany error: %v", err)
		return err
	}

	return nil
}

func GetWisers(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.Wiser, int64, error) {
	collection := mongodb.Collection(config.WiserTableName)

	var result []schema.Wiser
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, findOpt)
	if err != nil {
		return result, 0, err
	}
	defer cursor.Close(ctx)

	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	totalCount, err := collection.CountDocuments(ctx, filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetWisers ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}
