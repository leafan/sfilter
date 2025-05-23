package models

import (
	"context"
	"sfilter/config"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LoginHistoryModel struct {
	Collection *mongo.Collection
}

func (m *LoginHistoryModel) GetHistoriesByUsername(username string) ([]LoginHistory, error) {
	filter := bson.M{"username": username}
	options := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(10)

	var result []LoginHistory
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := m.Collection.Find(ctx, filter, options)
	if err != nil {
		return result, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &result)
	return result, err
}

func (m *LoginHistoryModel) CreatOne(h *LoginHistory) error {
	h.CreatedAt = time.Now()

	_, err := m.Collection.InsertOne(context.Background(), h)
	if err != nil {
		utils.Errorf("[ LoginHistoryModel CreatOne ] InsertOne error: %v, verifycode: %v\n", err, h)
	}

	return err
}
