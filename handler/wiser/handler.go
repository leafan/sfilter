package handler

import (
	"context"
	"sfilter/config"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewWiser(account, token, db string) *Wiser {
	clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
	mongodb, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		utils.Fatalf("connect mongo error: ", err)
	}

	wiserConfig := config.DefaultWiserConfig
	if account != "" {
		wiserConfig.DebugAccount = account
	}
	if token != "" {
		wiserConfig.DebugToken = token
	}
	if db != "" {
		config.DatabaseName = db
	}

	// other config change..

	return &Wiser{
		DB:     mongodb,
		Config: wiserConfig,
	}
}
