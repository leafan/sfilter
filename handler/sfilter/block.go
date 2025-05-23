package handler

import (
	"context"
	"errors"
	"math/big"

	"sfilter/config"
	"sfilter/schema"
	service_block "sfilter/services/block"
	"sfilter/services/chain"
	"sfilter/utils"

	"time"

	"github.com/cloudfresco/ethblocks"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

func getBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client, ethPrice float64) (*schema.Block, error) {
	if service_block.IsBlockProceeded(blockNumber.Int64(), mongodb) {
		utils.Warnf("[ getBlock ] Block is proceeded number: %v", blockNumber)
		return nil, errors.New("proceeded")
	}

	start := time.Now()
	ctx := context.Background()
	var err error

	oneBlk := new(schema.Block)

	if ethPrice == 0 {
		if config.DevelopmentMode {
			ethPrice, err = chain.GetBasicCoinPrice(client, blockNumber, config.BlockChain)
		} else {
			// 直接通过本地链上读取, 不要走 infura
			ethPrice, err = chain.GetBasicCoinPrice(client, nil, config.BlockChain)
		}

		if err != nil {
			return nil, err
		}
	}
	oneBlk.EthPrice = ethPrice

	block, err := ethblocks.GetBlockByNumber(ctx, client, blockNumber)
	if err != nil {
		utils.Errorf("[ getBlock ] GetBlockByNumber(%v) error: %v", blockNumber, err)
		return nil, err
	}

	oneBlk.Block = block
	utils.Debugf("Get block: %d now, tx num: %d, hash: %v, get txs time consumed: %v\n", blockNumber, len(block.Transactions()), block.Hash(), time.Since(start))

	for _, tx := range block.Transactions() {
		// 如果是单纯的transfer没有内容, 就pass掉
		if len(tx.Data()) <= 0 {
			// 单纯的transfer, continue掉
			continue
		}
		// utils.Debugf("[ getBlock ] tx: %v", tx.Hash())

		receipt, err := ethblocks.GetTransactionReceipt(ctx, client, tx.Hash())
		if err != nil {
			utils.Errorf("[ getBlock ] GetTransactionReceipt: %v err: %v", tx.Hash(), err)
			continue
		}

		oneTx := new(schema.Transaction)
		oneTx.OriginTx = tx
		oneTx.Receipt = receipt

		oneBlk.Transactions = append(oneBlk.Transactions, oneTx)
	}

	oneBlk.BlockNo = block.NumberU64()
	oneBlk.BlockTime = time.Unix(int64(block.Time()), 0)

	// schema.PrintOneBlock(oneBlk)

	utils.Debugf("[ getBlock ] get block: %d finished, valid tx num: %v, get logs time elapsed: % v\n", blockNumber, len(oneBlk.Transactions), time.Since(start))

	return oneBlk, nil
}
