package swap

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

func SaveSwapTx(swap *schema.Swap, mongodb *mongo.Client) error {
	collection := mongodb.Database(config.DatabaseName).Collection(config.SwapTableName)

	swap.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), swap)

	if err != nil && config.DevelopmentMode {
		// 如果是重复键值导致的插入失败, 且当前为修复模式, 则直接update
		filter := bson.D{{Key: "logIndexWithTx", Value: swap.LogIndexWithTx}}
		opts := options.Update().SetUpsert(true)

		update := bson.D{
			{Key: "$set", Value: swap},
		}
		_, err := collection.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			utils.Warnf("[ SaveSwapTx ] failed. hash: %v, err: %v\n", swap.TxHash, err)
		}
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
