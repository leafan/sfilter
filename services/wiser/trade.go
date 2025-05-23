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

func GetBiTrades(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.BiTrade, int64, error) {
	collection := mongodb.Collection(config.BiTradeTableName)

	var result []schema.BiTrade
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
		utils.Warnf("[ GetBiTrades ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}

func GetToBeSoldBiTrade(mainToken string, mongodb *mongo.Client) (*schema.BiTrade, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiTradeTableName)

	filter := bson.D{
		{Key: "mainToken", Value: mainToken},
		{Key: "status", Value: 0},
	}

	var result schema.BiTrade
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		// utils.Warnf("[ GetToBeSoldTrade ] didn't find mainToken: %v", mainToken)
		return nil, err
	}

	return &result, nil
}

func SaveBiTrade(trade *schema.BiTrade, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiTradeTableName)

	trade.UpdatedAt = time.Now()
	trade.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), trade)
	if err != nil {
		utils.Errorf("[ SaveTrade ] failed. trade: %v, err: %v\n", trade, err)
	}
}

func UpdateBiTrade(trade *schema.BiTrade, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiTradeTableName)

	filter := bson.D{
		{Key: "pairAddress", Value: trade.PairAddress},
		{Key: "status", Value: 0},
	}

	update := bson.D{
		{Key: "$set", Value: trade},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Errorf("[ ReplaceOne ] failed. trade: %v, err: %v\n", trade, err)
	}
}
