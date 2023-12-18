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

func SaveSwapTx(swap *schema.Swap, mongodb *mongo.Client) error {
	collection := mongodb.Database(config.DatabaseName).Collection(config.SwapTableName)

	swap.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), swap)
	if err != nil {
		utils.Warnf("[ saveSwapTx ] InsertOne error: %v, swap tx: %v\n", err, swap.TxHash)
	}

	return err
}

func GetSwapEvents(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.Swap, int64, error) {
	collection := mongodb.Collection(config.SwapTableName)

	var result []schema.Swap
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
		utils.Warnf("[ GetSwapEvents ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}
