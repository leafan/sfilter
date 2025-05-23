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

func GetTrendByKey(key string, mongodb *mongo.Client) (*schema.GlobalTrend, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.GlobalTrendTableName)
	filter := bson.D{{Key: "timelineKey", Value: key}}

	var result schema.GlobalTrend
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	return &result, err
}

func GetTrendsByTimeRange(start, end time.Time, mongodb *mongo.Client) ([]schema.GlobalTrend, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.GlobalTrendTableName)

	filter := bson.D{
		{Key: "timestamp", Value: bson.D{
			{Key: "$gte", Value: start},
			{Key: "$lte", Value: end},
		}},
	}
	options := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	var result []schema.GlobalTrend
	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		utils.Warnf("[ GetTrendsByTimeRange ] Find error: %v, filter: %v", err, filter)

		return result, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &result); err != nil {
		utils.Warnf("[ GetTrendsByTimeRange ] cursor.All error: %v, filter: %v\n", err, filter)
	}

	return result, err
}

func UpsertTrends(trends *schema.GlobalTrend, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.GlobalTrendTableName)
	filter := bson.D{{Key: "timelineKey", Value: trends.TimelineKey}}
	opt := options.Update().SetUpsert(true)

	if trends.CreatedAt.IsZero() {
		trends.CreatedAt = time.Now()
	}

	update := bson.D{
		{Key: "$set", Value: trends},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ UpsertTrends ] failed. key: %v, err: %v\n", trends.TimelineKey, err)
	}
}
