package wiser

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

func GetBiDeals(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.BiDeal, int64, error) {
	collection := mongodb.Collection(config.BiDealTableName)

	var result []schema.BiDeal
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
		utils.Warnf("[ GetBiDeals ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}

func SaveDeal(deal *schema.BiDeal, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiDealTableName)

	deal.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), deal)
	if err != nil {
		// 这里失败是正常现象, 因为可能重复计算导致
		utils.Warnf("[ SaveDeals ] InsertOne error: %v, deal: %v\n", err, deal)
	}
}
