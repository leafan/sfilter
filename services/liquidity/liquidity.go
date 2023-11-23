package liquidity

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func SaveLiquidityEvent(event *schema.LiquidityEvent, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.LiquidityEventTableName)

	event.CreatedAt = time.Now()

	collection.InsertOne(context.Background(), event)
}
