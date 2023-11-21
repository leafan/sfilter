package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cloudfresco/ethblocks"

	"sfilter/config"
	"sfilter/handler"
	"sfilter/schema"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func test() {
	fmt.Printf("\n\n\t****** debug start ******\n\n")

	// chain.TEST_PAIR()
	// chain.TEST_TOKEN()
	// kline.TEST_KLINE()
	// handler.TEST_HANDLER()

	// kline.TEST_KLINE_DB()

	fmt.Printf("\t****** debug end ******\n\n\n\n")

	// os.Exit(0)
}

func main() {
	test()

	client, mongodb := _init()
	loop(client, mongodb)
}

func _init() (*ethclient.Client, *mongo.Client) {
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
	log.Printf("[ loop ] start SubscribeNewHead now..\n\n")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal("SubscribeBlocks error: ", err)
			return

		case header := <-headers:
			log.Printf("[ loop ] get new header now. number: %v\n\n\n", header.Number)
			go handler.HandleBlock(header.Number, client, mongodb)
		}

	}
}
