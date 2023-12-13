package main

import (
	"context"
	"flag"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cloudfresco/ethblocks"

	"sfilter/config"
	"sfilter/handler"
	"sfilter/schema"
	"sfilter/user/auth"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func test() {
	utils.Infof("****** Debug start ******\n\n")

	// chain.TEST_POOL()
	auth.TEST_EMAIL()

	utils.Infof("****** Debug end  ******\n\n\n")
}

func main() {
	test()

	db := flag.String("db", "", "the db want to use")
	block := flag.Int64("block", 0, "a tool to retrive/test one block")
	flag.Parse()

	client, mongodb := _init(*db)

	if *block != 0 {
		utils.Infof("\n\n\nStart block test now...\n\n\n")
		handler.HandleBlock(big.NewInt(*block), client, mongodb)
		utils.Infof("\n\nFinished block test...\n\n")

		return
	}

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
		panic(err)
	}
	utils.Infof("[ loop ] start SubscribeNewHead now..\n\n")

	for {
		select {
		case err := <-sub.Err():
			utils.Fatalf("SubscribeBlocks error: ", err)
			return

		case header := <-headers:
			utils.Infof("[ loop ] Get new header now. number: %v\n", header.Number)
			go handler.HandleBlock(header.Number, client, mongodb)
		}

	}
}
