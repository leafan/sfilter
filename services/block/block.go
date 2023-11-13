package block

import (
	"context"
	"log"
	"math/big"
	"time"

	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/swap"

	"github.com/cloudfresco/ethblocks"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client, ethPrice float64) {
	if isBlockProceeded(blockNumber.Int64(), mongodb) {
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

	if ethPrice == 0 {
		ethPrice = chain.GetEthPrice(client, blockNumber)
	}
	oneBlk.EthPrice = ethPrice

	// schema.PrintOneBlock(oneBlk)

	go doHandleBlock(oneBlk, mongodb)

	bps := &schema.BlockProceeded{
		BlockNo:   blockNumber.Int64(),
		Hash:      block.Hash().String(),
		BlockTime: block.Time(),

		EthPrice: ethPrice,

		CreatedAt: time.Now(),
	}

	saveBlockProceeded(bps, mongodb)

	log.Printf("[ getBlock ] finished block: %d, time elapsed: % v\n\n", blockNumber, time.Since(start))
}

func doHandleBlock(blk *schema.Block, mongodb *mongo.Client) {
	go swap.Swap_handler(blk, mongodb)

	// todo
	// go Transfer_handler(blk, mongodb)
}

func isBlockProceeded(blkNo int64, mongodb *mongo.Client) bool {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	filter := bson.M{"blockno": blkNo}

	var result bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		// log.Printf("[ isBlockProceeded ] Find block %v failed. err: %v\n", blkNo, err)
		return false
	}

	if len(result) == 0 {
		// log.Println("[ isBlockProceeded ] FindOne result 0...")
		return false
	}

	// log.Println("[ isBlockProceeded ] FindOne result 1...")
	return true
}

func saveBlockProceeded(bps *schema.BlockProceeded, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	_, err := collection.InsertOne(context.Background(), bps)
	if err != nil {
		log.Printf("[ SaveBlockProceeded ] InsertOne error: %v, block no: %v\n", err, bps.BlockNo)
		return
	}

}
