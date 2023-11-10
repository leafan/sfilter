/**
* 优秀参考对象:
 1. https://github.com/mellaught/ethereum-blocks/blob/master/src/ethereum/blocks.go
 2. https://github.com/Orochyy/blockchainETH-MongoDb/blob/main/modules/main.go
*/

package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cloudfresco/ethblocks"

	"sfilter/handler"
	"sfilter/schema"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ws_addr = "ws://127.0.0.1:8546"
var mongo_addr = "mongodb://127.0.0.1:27017"

func main() {
	// todo
	// retrive_old_blocks()

	loop()
}

func loop() {
	client, err := ethblocks.GetClient(ws_addr)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		panic(err)
	}
	log.Printf("[ loop ] start SubscribeNewHead now..\n\n")

	clientOptions := options.Client().ApplyURI(mongo_addr)
	mongodb, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = mongodb.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		select {
		case err := <-sub.Err():
			log.Fatal("SubscribeBlocks error: ", err)
			return

		case header := <-headers:
			log.Printf("\n\n\n\n\n[ loop ] get new header now. number: %v\n\n\n", header.Number)
			go getBlock(header.Number, client, mongodb)
		}

	}
}

func getBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client) {
	start := time.Now()
	ctx := context.Background()

	oneBlk := new(schema.Block)

	block, err := ethblocks.GetBlockByNumber(ctx, client, blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	oneBlk.Block = block

	// ethblocks.PrintTransaction(txs[0])
	for _, tx := range block.Transactions() {
		receipt, err := ethblocks.GetTransactionReceipt(ctx, client, tx.Hash())
		if err != nil {
			log.Println("[ getBlock ] GetTransactionReceipt err: ", err)
			continue
		}

		oneTx := new(schema.Transaction)
		oneTx.OriginTx = tx
		oneTx.Receipt = receipt

		oneBlk.Transactions = append(oneBlk.Transactions, oneTx)
	}

	// schema.PrintOneBlock(oneBlk)

	go handleBlock(oneBlk, mongodb)

	log.Printf("[ getBlock ] finished block: %d, time elapsed: % v\n\n", blockNumber, time.Since(start))
}

func handleBlock(blk *schema.Block, mongodb *mongo.Client) {
	go handler.Swap_handle(blk, mongodb)
}
