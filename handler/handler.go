package handler

import (
	"context"
	"math/big"
	"sync"
	"time"

	"sfilter/config"
	"sfilter/handler/facet"
	"sfilter/schema"
	service_block "sfilter/services/block"
	"sfilter/services/chain"
	"sfilter/utils"

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
		utils.Fatalf("[ Retrive_old_blocks ] HeaderByNumber err: ", err)
	}

	startBlock := curBlkNo.Number.Int64() - int64(config.RetriveOldBlockNum)
	utils.Infof("[ Retrive_old_blocks ] retrive now.. start block: %v", startBlock)

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

	// 获取transfer信息
	transferMapForSwap, transfers := HandleTransfer(blk, mongodb)

	// 获取swaps
	swaps := HandleSwapAndKline(blk, transferMapForSwap, mongodb)

	// 更新transfer的usd value等
	UpSaveTransferInfoBySwaps(transfers, swaps, mongodb)

	// trade info 是更新最近24h或7天的数据, 因此老数据就别掺和了
	if time.Since(time.Unix(int64(blk.Block.Time()), 0)).Seconds() < config.SecondsForOneWeek {
		HandleTradeInfo(blk, mongodb, swaps)
		HandleGlobalInfo(blk, mongodb)
	}

	// 更新 用户跟踪地址逻辑
	HandleUserTrackSwaps(blk, mongodb, swaps)

	// 更新 facet 逻辑
	facet.HandleFacetLogic(blk, mongodb)

	// etc.. todo

	// record the proceeded block.
	setBlockToProceeded(blk, mongodb)

	utils.Debugf("Handle block: %d finished, swap num: %v, time elapsed: % v\n", blk.Block.NumberU64(), blk.TxNums, time.Since(start))
}

func setBlockToProceeded(block *schema.Block, mongodb *mongo.Client) {
	bps := &schema.BlockProceeded{
		BlockNo:     block.Block.Number().Int64(),
		Hash:        block.Block.Hash().String(),
		BlockTime:   int64(block.Block.Time()),
		TxNums:      block.TxNums,
		VolumeByUsd: block.VolumeByUsd,

		EthPrice: block.EthPrice,
	}

	service_block.SetBlockProceeded(bps, mongodb)

}

func TEST_HANDLER() {
	blockNo := int64(18615952)
	HandleBlock(big.NewInt(blockNo), chain.GetEthClient(), chain.GetMongo())
}
