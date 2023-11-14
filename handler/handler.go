package handler

import (
	"math/big"

	"sfilter/schema"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client) {
	block := getBlock(blockNumber, client, mongodb, 0)

	if block != nil {
		handleOneBlock(block, mongodb)
	}
}

func handleOneBlock(blk *schema.Block, mongodb *mongo.Client) {
	go HandleSwap(blk, mongodb)

	// todo
	// go Transfer_handler(blk, mongodb)
}
