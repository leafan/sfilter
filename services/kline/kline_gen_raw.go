package kline

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

// 取出从start到end的所有柱子
func get1MinKlineByPair(pair string, start, end time.Time, mongodb *mongo.Database) []schema.KLines1Min {
	collection := mongodb.Collection(config.Kline1MinTableName)

	filter := bson.M{
		"pair": pair,
		"timestamp": bson.M{
			"$gte": start, "$lt": end,
		},
	}
	options := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	// log.Printf("[ get1MinKlineByPair ] debug filter: %v\n", filter)

	var result []schema.KLines1Min
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			utils.Warnf("[ get1MinKlineByPair ] Find error: %v, filter: %v\n", err, filter)
		}

		return result
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &result); err != nil {
		utils.Warnf("[ get1MinKlineByPair ] cursor.All error: %v, filter: %v\n", err, filter)
	}
	// log.Printf("[ get1MinKlineByPair ] result: %v\n\n", result)

	return result
}

func get1HourKlineByPair(pair string, start, end time.Time, mongodb *mongo.Database) []schema.KLines1Hour {
	collection := mongodb.Collection(config.Kline1HourTableName)

	filter := bson.M{
		"pair": pair,
		"timestamp": bson.M{
			"$gte": start, "$lt": end,
		},
	}
	options := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	var result []schema.KLines1Hour
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			utils.Warnf("[ get1HourKlineByPair ] Find error: %v, filter: %v\n", err, filter)
		}

		return result
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &result); err != nil {
		utils.Warnf("[ get1HourKlineByPair ] cursor.All error: %v, filter: %v\n", err, filter)
	}

	return result
}
