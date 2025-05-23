package utils

import (
	"sfilter/config"

	"go.mongodb.org/mongo-driver/mongo"
)

var mongodb *mongo.Client

func InitMongo(db *mongo.Client) {
	mongodb = db
}

// func GetMongo() *mongo.Client {
// 	return mongodb
// }

func GetChainDatabase(chain string) *mongo.Database {
	var dbName string

	switch chain {
	case "avax":
		dbName = "avax"

	case "optimism":
		dbName = "optimism"

	default:
		dbName = config.DatabaseName
	}

	return mongodb.Database(dbName)
}
