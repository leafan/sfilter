package handler

import (
	"context"
	"log"
	"math/big"
	"time"

	"sfilter/schema"
	services_block "sfilter/services/block"

	"github.com/cloudfresco/ethblocks"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client) {
	if services_block.IsBlockProceeded(blockNumber.Int64(), mongodb) {
		return
	}

	start := time.Now()
	ctx := context.Background()

	oneBlk := new(schema.Block)

	block, err := ethblocks.GetBlockByNumber(ctx, client, blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	oneBlk.Block = block

	log.Printf("\n[ getBlock ] get block: %d now, tx num: %d, hash: %v\n", blockNumber, len(block.Transactions()), block.Hash())

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

	go doHandleBlock(oneBlk, mongodb)

	bps := &schema.BlockProceeded{
		BlockNo: blockNumber.Int64(),
		Hash:    block.Hash().String(),

		CreatedAt: time.Now(),
	}

	services_block.SaveBlockProceeded(bps, mongodb)

	log.Printf("[ getBlock ] finished block: %d, time elapsed: % v\n\n", blockNumber, time.Since(start))
}

func doHandleBlock(blk *schema.Block, mongodb *mongo.Client) {
	go Swap_handler(blk, mongodb)

	// todo
	// go Transfer_handler(blk, mongodb)
}
