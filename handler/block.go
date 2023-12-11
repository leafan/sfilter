package handler

import (
	"context"
	"math/big"

	"sfilter/schema"
	service_block "sfilter/services/block"
	"sfilter/services/chain"
	"sfilter/utils"

	"time"

	"github.com/cloudfresco/ethblocks"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

func getBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client, ethPrice float64) *schema.Block {
	if service_block.IsBlockProceeded(blockNumber.Int64(), mongodb) {
		utils.Warnf("[ getBlock ] Block is proceeded number: ", blockNumber)
		return nil
	}

	start := time.Now()
	ctx := context.Background()

	oneBlk := new(schema.Block)

	block, err := ethblocks.GetBlockByNumber(ctx, client, blockNumber)
	if err != nil {
		utils.Errorf("[ getBlock ] GetBlockByNumber error: ", err)
		return nil
	}
	oneBlk.Block = block
	utils.Debugf("Get block: %d now, tx num: %d, hash: %v\n", blockNumber, len(block.Transactions()), block.Hash())

	for _, tx := range block.Transactions() {
		receipt, err := ethblocks.GetTransactionReceipt(ctx, client, tx.Hash())
		if err != nil {
			utils.Errorf("[ getBlock ] GetTransactionReceipt err: ", err)
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

	oneBlk.BlockNo = block.NumberU64()
	oneBlk.BlockTime = time.Unix(int64(block.Time()), 0)

	// schema.PrintOneBlock(oneBlk)

	utils.Debugf("[ getBlock ] get block: %d finished, time elapsed: % v\n", blockNumber, time.Since(start))

	return oneBlk
}
