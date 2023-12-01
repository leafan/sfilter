package liquidity

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

func GetLiquidityEvents(page, limit int64, mongodb *mongo.Client) ([]schema.LiquidityEvent, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.LiquidityEventTableName)

	skip := int64(page*limit - limit)

	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	var result []schema.LiquidityEvent
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{}, options)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("[ GetLiquidityEvents ] Find error: %v\n", err)
		}

		return result, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &result)
	return result, err
}

func SaveLiquidityEvent(event *schema.LiquidityEvent, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.LiquidityEventTableName)

	event.CreatedAt = time.Now()

	collection.InsertOne(context.Background(), event)
}
