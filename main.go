/**
* 优秀参考对象:
 1. https://github.com/mellaught/ethereum-blocks/blob/master/src/ethereum/blocks.go
 2. https://github.com/Orochyy/blockchainETH-MongoDb/blob/main/modules/main.go
*/

package main

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cloudfresco/ethblocks"

	"sfilter/config"
	"sfilter/handler"
	"sfilter/schema"
	"sfilter/services/chain"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func test() {
	chain.TEST_PAIR()
	// chain.TEST_TOKEN()

	// os.Exit(0)
}

func main() {
	client, mongodb := _init()

	// test()

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
