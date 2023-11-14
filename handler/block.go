package handler

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"sfilter/config"
	"sfilter/schema"
	service_block "sfilter/services/block"
	"sfilter/services/chain"

	"time"

	"github.com/cloudfresco/ethblocks"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

// 每次启动往回回溯n个区块, 防止某一次未处理
// 回溯的时候, eth价格通过infura获取, 每10个区块更新一次价格
func Retrive_old_blocks(client *ethclient.Client, mongodb *mongo.Client) {
	curBlkNo, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal("[ Retrive_old_blocks ] HeaderByNumber err: ", err)
	}

	startBlock := curBlkNo.Number.Int64() - int64(config.RetriveOldBlockNum)

	ethPrice := chain.GetEthPrice(client, big.NewInt(startBlock))
	for i := startBlock; i < curBlkNo.Number.Int64(); i++ {
		if i%10 == 0 {
			fmt.Println("Retrive_old_blocks debug i: ", i)
			ethPrice = chain.GetEthPrice(client, big.NewInt(i))
		}

		go func(i int64, ethPrice float64) {
			block := getBlock(big.NewInt(i), client, mongodb, ethPrice)
			if block != nil {
				handleOneBlock(block, mongodb)
			}
		}(i, ethPrice)

		time.Sleep(20 * time.Millisecond)
	}

}

func getBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client, ethPrice float64) *schema.Block {
	if service_block.IsBlockProceeded(blockNumber.Int64(), mongodb) {
		log.Println("[ getBlock ] Block is proceeded number: ", blockNumber)
		return nil
	}

	start := time.Now()
	ctx := context.Background()

	oneBlk := new(schema.Block)

	block, err := ethblocks.GetBlockByNumber(ctx, client, blockNumber)
	if err != nil {
		log.Println("[ getBlock ] GetBlockByNumber error: ", err)
		return nil
	}
	oneBlk.Block = block

	log.Printf("\n[ getBlock ] block number: %d now, tx num: %d, hash: %v\n", blockNumber, len(block.Transactions()), block.Hash())

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

	bps := &schema.BlockProceeded{
		BlockNo:   blockNumber.Int64(),
		Hash:      block.Hash().String(),
		BlockTime: block.Time(),

		EthPrice: ethPrice,

		CreatedAt: time.Now(),
	}

	service_block.SaveBlockProceeded(bps, mongodb)

	log.Printf("[ getBlock ] finished block: %d, time elapsed: % v\n\n", blockNumber, time.Since(start))

	return oneBlk
}
