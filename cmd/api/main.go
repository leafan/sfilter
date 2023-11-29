package main

import (
	"context"
	"log"
	"sfilter/api"
	"sfilter/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
	mongodb, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.New()

	server := api.NewServer(r, mongodb)
	server.Route()

	server.Run(config.ApiListenAddrPort)
}
