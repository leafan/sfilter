package main

import (
	"context"
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cloudfresco/ethblocks"

	"sfilter/config"
	handler "sfilter/handler/sfilter"
	"sfilter/schema"
	userModels "sfilter/user/models"
	"sfilter/utils"

	sblock "sfilter/services/block"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	db := flag.String("db", "", "the db want to use")
	block := flag.Int64("block", 0, "a tool to retrive/test one block")
	flag.Parse()

	client, mongodb := _init(*db)

	if *block != 0 {
		utils.Infof("\n\n\nStart block test now...\n\n\n")

		config.DevelopmentMode = true

		// 先把block id set 未处理
		sblock.SetUnProceeded(*block, mongodb)
		handler.HandleBlock(big.NewInt(*block), client, mongodb)

		utils.Infof("\n\nFinished block test...\n\n")
		return
	}

	go getTrackAddressOnTimer(mongodb)

	// 主循环, 不退出
	loop(client, mongodb)
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

func loop(client *ethclient.Client, mongodb *mongo.Client) {
	go handler.Retrive_old_blocks(client, mongodb) // 先回溯

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		utils.Fatalf("SubscribeNewHead error: %v", err)
	}
	utils.Infof("[ loop ] start SubscribeNewHead now..\n\n")

	for {
		select {
		case err := <-sub.Err():
			utils.Fatalf("SubscribeBlocks error: %v", err)

		case header := <-headers:
			utils.Infof("[ loop ] Get new header now. number: %v\n", header.Number)
			go handler.HandleBlock(header.Number, client, mongodb)
		}

	}
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
