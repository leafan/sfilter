package token

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	filter := bson.D{{Key: "address", Value: token.Address}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "updatedat", Value: time.Now()},
		}},
	}

	token.UpdatedAt = time.Now()

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("[ SaveTokenInfo ] UpdateOne error: %v, token: %v\n", err, token.Address)
		return
	} else {
		log.Printf("[ SaveTokenInfo ] UpdateOne success. token: %v\n", token.Address)
	}

}
