package main

import (
	"context"
	"sfilter/api"
	"sfilter/config"
	"sfilter/user"
	"sfilter/utils"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
	mongodb, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		utils.Fatalf("connect mongo error: ", err)
	}

	r := gin.New()

	initUserServer(r)

	server := api.NewServer(r, mongodb)
	server.Route()

	utils.Debugf("run server now...")
	server.Run(config.ApiListenAddrPort)
}

func initUserServer(r *gin.Engine) {
	user.Run(r)
}
