package swap

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveTrackSwap(uswap *schema.TrackSwap, mongodb *mongo.Client) error {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TrackSwapTableName)

	uswap.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), uswap)
	if err != nil {
		utils.Warnf("[ SaveTrackSwap ] InsertOne error: %v, swap tx: %v\n", err, uswap.TxHash)
	}

	return err
}

func GetTrackSwaps(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.TrackSwap, int64, error) {
	collection := mongodb.Collection(config.TrackSwapTableName)

	var result []schema.TrackSwap
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
		utils.Warnf("[ GetTrackSwaps ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}
