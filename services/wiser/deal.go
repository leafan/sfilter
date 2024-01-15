package wiser

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func SaveDeal(deal *schema.BiDeal, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiDealTableName)

	deal.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), deal)
	if err != nil {
		// 这里失败是正常现象, 因为可能重复计算导致
		utils.Warnf("[ SaveDeals ] InsertOne error: %v, deal: %v\n", err, deal)
	}
}
