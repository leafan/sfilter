package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"

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
	fmt.Printf("\n\n****** \033[0;34mDebug start \033[0m ******\n\n")

	// chain.TEST_POOL()

	fmt.Printf("****** \033[0;34mDebug end \033[0m ******\n\n\n")
}

func main() {
	test()

	db := flag.String("db", "", "the db want to use")
	block := flag.Int64("block", 0, "a tool to retrive/test one block")
	flag.Parse()

	client, mongodb := _init(*db)

	if *block != 0 {
		fmt.Printf("\n\n\n\033[0;34mStart block test now...\033[0m\n\n\n")
		handler.HandleBlock(big.NewInt(*block), client, mongodb)
		fmt.Printf("\n\n\033[0;34mFinished block test...\033[0m\n\n")
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
	log.Printf("[ loop ] start SubscribeNewHead now..\n\n")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal("SubscribeBlocks error: ", err)
			return

		case header := <-headers:
			log.Printf("\033[0;34m[ loop ] get new header now. number: %v\033[0m\n\n", header.Number)
			go handler.HandleBlock(header.Number, client, mongodb)
		}

	}
}
