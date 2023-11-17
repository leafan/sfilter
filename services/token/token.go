package token

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveTokenInfo(token *schema.Token, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), token)
	if err != nil {
		log.Printf("[ SaveTokenInfo ] InsertOne error: %v, token: %v\n", err, token.Address)
		return
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

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		log.Printf("[ UpdateTokenInfo ] UpdateOne error: %v, token: %v\n", err, token.Address)
		return
	} else {
		log.Printf("[ UpdateTokenInfo ] UpdateOne success. token: %v\n", token.Address)
	}

}
