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

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cloudfresco/ethblocks"

	"sfilter/schem"
)

var ws_addr = "ws://127.0.0.1:8546"

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

	log.Println("[ loop ] start SubscribeNewHead now..\n\n")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal("SubscribeBlocks error: ", err)
			return

		case header := <-headers:
			log.Println("\n[ loop ] get new header now. number: ", header.Number)
			go getBlock(header.Number, client)
		}

	}
}

func getBlock(blockNumber *big.Int, client *ethclient.Client) {
	ctx := context.Background()

	oneBlk := new(schem.OneBlock)

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

		oneTx := new(schem.OneTransaction)
		oneTx.OriginTx = tx
		oneTx.Receipt = receipt

		oneBlk.Transactions = append(oneBlk.Transactions, oneTx)
	}

	// schem.PrintOneBlock(oneBlk)

	// go handleBlock()

	log.Printf("\n\n[ getBlock ] finished block: %d\n\n", blockNumber)
}
