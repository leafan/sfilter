package utils

import "go.mongodb.org/mongo-driver/mongo"

var mongodb *mongo.Client

func InitMongo(db *mongo.Client) {
	mongodb = db
}

func GetMongo() *mongo.Client {
	return mongodb
}
