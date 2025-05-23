package handler

import (
	"sfilter/schema"
	"sfilter/services/kline"

	"go.mongodb.org/mongo-driver/mongo"
)

func UpdateKlines(swap *schema.Swap, mongodb *mongo.Client) {
	kline.Update1MinKline(swap, mongodb)
	kline.Update1HourKline(swap, mongodb)

	// update1DayKline(swap, mongodb)
}
