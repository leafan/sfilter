package liquidity

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

func GetLiquidityEvents(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Client) ([]schema.LiquidityEvent, int64, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.LiquidityEventTableName)

	var result []schema.LiquidityEvent
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, findOpt)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			utils.Warnf("[ GetLiquidityEvents ] Find error: %v\n", err)
		}

		return result, 0, err
	}
	defer cursor.Close(ctx)

	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	totalCount, err := collection.CountDocuments(ctx, filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetLiquidityEvents ] Count error: %v\n", err)
		return result, 0, err // 返回总数为0
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}

func SaveLiquidityEvent(event *schema.LiquidityEvent, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.LiquidityEventTableName)

	event.CreatedAt = time.Now()

	collection.InsertOne(context.Background(), event)
}
