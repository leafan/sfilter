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

func SaveUpsert1MinKline(kline *schema.KLines1Min, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1MinTableName)

	kline.UpdatedAt = time.Now()

	filter := bson.D{{Key: "pairDayHour", Value: kline.PairDayHour}}
	update := bson.D{{Key: "$set", Value: kline}}
	opt := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ SaveUpsert1MinKline ] InsertOne error: %v, kline: %v\n", err, kline)
		return
	}
}

func SaveUpsert1HourKline(kline *schema.KLines1Hour, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1HourTableName)

	kline.UpdatedAt = time.Now()

	filter := bson.D{{Key: "pairMonthDay", Value: kline.PairMonthDay}}
	update := bson.D{{Key: "$set", Value: kline}}
	opt := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ SaveUpsert1HourKline ] InsertOne error: %v, kline: %v\n", err, kline)
		return
	}
}
