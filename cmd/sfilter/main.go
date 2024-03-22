package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cloudfresco/ethblocks"

	"sfilter/config"
	handler "sfilter/handler/sfilter"
	"sfilter/schema"
	userModels "sfilter/user/models"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	db := flag.String("db", "", "the db want to use")
	block := flag.Int64("block", 0, "a tool to retrive/test one block")
	flag.Parse()

	client, mongodb := _init(*db)

	h, err := handler.NewHandler(client, mongodb)
	if err != nil {
		utils.Fatalf("[ loop ] NewHandler failed: %v", err)
	}

	// go getTrackAddressOnTimer(mongodb)

	h.Run(*block)

}

func _init(db string) (*ethclient.Client, *mongo.Client) {
	if db != "" {
		config.DatabaseName = db
	}

	client, err := ethblocks.GetClient(config.WS_ADDR)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
	mongodb, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	schema.InitTables(mongodb)

	return client, mongodb
}

func getTrackAddressOnTimer(mongodb *mongo.Client) {
	userModels.InitService(mongodb) // 初始化db, 可以直接使用user的service等

	// 第一次执行
	err := handler.GetTrackAddressMapOnTimer()
	if err != nil {
		utils.Fatalf("[ getTrackAddressOnTimer ] GetTrackAddressOnTimer failed, err:  %v", err)
	}

	go func() {
		updateTimer := time.NewTicker(config.UpdateTrackAddressInterval)
		defer updateTimer.Stop()

		for range updateTimer.C {
			handler.GetTrackAddressMapOnTimer()
		}
	}()

	go func() {
		checkTimer := time.NewTicker(config.CheckTrackAddressInterval)
		defer checkTimer.Stop()

		for range checkTimer.C {
			handler.CheckAndDeleteUserSwapsUpLimit(mongodb)
		}
	}()
}
