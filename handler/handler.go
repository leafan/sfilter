package handler

import (
	"context"
	"log"
	"math/big"
	"sync"
	"time"

	"sfilter/config"
	"sfilter/schema"
	service_block "sfilter/services/block"
	"sfilter/services/chain"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client) {
	block := getBlock(blockNumber, client, mongodb, 0)

	if block != nil {
		handleOneBlock(block, mongodb)
	}
}

// 每次启动往回回溯n个区块, 防止某一次未处理
// 回溯的时候, eth价格通过infura获取, 每10个区块更新一次价格
func Retrive_old_blocks(client *ethclient.Client, mongodb *mongo.Client) {
	curBlkNo, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal("[ Retrive_old_blocks ] HeaderByNumber err: ", err)
	}

	startBlock := curBlkNo.Number.Int64() - int64(config.RetriveOldBlockNum)

	var ethPrice float64
	times := 0

	// 控制并发执行的协程数量
	ch := make(chan struct{}, config.MaxConcurrentRoutineNums)
	var wg sync.WaitGroup

	for i := startBlock; i < curBlkNo.Number.Int64(); i++ {
		if service_block.IsBlockProceeded(i, mongodb) {
			continue
		}

		if times%config.GetPriceIntervalForRetrive == 0 {
			ethPrice = chain.GetEthPrice(client, big.NewInt(i))
		}

		ch <- struct{}{} // 当协程处理不过来的时候, 这里会阻塞，写不进去
		wg.Add(1)

		go func(i int64, ethPrice float64) {
			defer wg.Done()

			block := getBlock(big.NewInt(i), client, mongodb, ethPrice)
			<-ch // 处理完了, 释放一个协程位置

			// 先释放协程, 再往后走, 因为这里面有 sleep
			if block != nil {
				handleOneBlock(block, mongodb)
			}

		}(i, ethPrice)

		times++

		time.Sleep(config.SleepIntervalforRetrive * time.Millisecond)
	}

	close(ch)
	wg.Wait()
}

// 串行处理即可, 因为跑到这里来的都是协程
func handleOneBlock(blk *schema.Block, mongodb *mongo.Client) {
	start := time.Now()

	HandlePairLogic(blk, mongodb)
	HandleLiquidityLogic(blk, mongodb)

	maps := HandleTransfer(blk, mongodb)

	swaps := HandleSwap(blk, maps, mongodb)
	blk.TxNums = len(swaps)

	if !config.GET_VERY_OLD_DATA_DEBUG {
		HandleTradeInfo(blk, mongodb, swaps)
	}

	// etc.. todo

	// record the proceeded block.
	setBlockToProceeded(blk, mongodb)

	log.Printf("[ handleOneBlock ] handle block: %d finished, swap num: %v, time elapsed: % v\n\n", blk.Block.NumberU64(), len(swaps), time.Since(start))
}

func setBlockToProceeded(block *schema.Block, mongodb *mongo.Client) {
	bps := &schema.BlockProceeded{
		BlockNo:   block.Block.Number().Int64(),
		Hash:      block.Block.Hash().String(),
		BlockTime: int64(block.Block.Time()),
		TxNums:    block.TxNums,

		EthPrice: block.EthPrice,
	}

	service_block.SaveBlockProceeded(bps, mongodb)

}

func TEST_HANDLER() {
	blockNo := int64(18615952)
	HandleBlock(big.NewInt(blockNo), chain.GetEthClient(), chain.GetMongo())
}
