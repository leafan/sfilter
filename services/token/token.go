package token

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tokenLock sync.Mutex

func GetTokenInfo(address string, mongodb *mongo.Client) (*schema.Token, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	filter := bson.D{{Key: "address", Value: address}}

	var result schema.Token
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	return &result, err
}

func SaveTokenInfo(token *schema.Token, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()

	tokenLock.Lock()
	defer tokenLock.Unlock()

	_, err := collection.InsertOne(context.Background(), token)
	if err != nil {
		utils.Warnf("[ SaveTokenInfo ] InsertOne error: %v, token: %v\n", err, token.Address)
		return
	}

}

// 更新token价格字段
func UpdateTokenPrice(address string, price float64, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	filter := bson.D{{Key: "address", Value: address}}
	opt := options.Update().SetUpsert(true)

	info := struct {
		PriceInUsd float64   `bson:"priceInUsd"`
		UpdatedAt  time.Time `bson:"updatedAt"`
	}{
		PriceInUsd: price,
		UpdatedAt:  time.Now(),
	}

	update := bson.M{
		"$set": info,
	}

	tokenLock.Lock()
	defer tokenLock.Unlock()

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ UpdateTokenPrice ] UpdateOne error: %v, token: %v\n", err, address)
	}
}

// 更新 token
func UpdateTokenInfo(token *schema.Token, mongodb *mongo.Client) {
	if token == nil {
		return
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	token.UpdatedAt = time.Now()

	filter := bson.D{{Key: "address", Value: token.Address}}
	opt := options.Update().SetUpsert(true)

	update := bson.M{
		"$set": token,
	}

	tokenLock.Lock()
	defer tokenLock.Unlock()

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ UpdateTokenInfo ] UpdateOne error: %v, token: %v\n", err, token.Address)
	}

}
