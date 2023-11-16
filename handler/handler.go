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

	ethPrice := chain.GetEthPrice(client, big.NewInt(startBlock))
	times := 0

	// 控制并发执行的协程数量
	ch := make(chan struct{}, config.MaxRoutineNums)
	var wg sync.WaitGroup

	for i := startBlock; i < curBlkNo.Number.Int64(); i++ {
		if service_block.IsBlockProceeded(i, mongodb) {
			continue
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
		if times%config.GetPriceIntervalForRetrive == 0 {
			ethPrice = chain.GetEthPrice(client, big.NewInt(i))
		}

		time.Sleep(config.SleepIntervalforRetrive * time.Millisecond)
	}

	close(ch)
	wg.Wait()
}

func handleOneBlock(blk *schema.Block, mongodb *mongo.Client) {
	go HandlePairCreated(blk, mongodb)

	go HandleSwap(blk, mongodb)

	// todo
	setBlockToProceeded(blk, mongodb)
}

func setBlockToProceeded(block *schema.Block, mongodb *mongo.Client) {
	time.Sleep(3 * time.Second) // sleep 几秒再标记处理完成，等待 block 处理完成

	bps := &schema.BlockProceeded{
		BlockNo:   block.Block.Number().Int64(),
		Hash:      block.Block.Hash().String(),
		BlockTime: int64(block.Block.Time()),

		EthPrice: block.EthPrice,
	}

	service_block.SaveBlockProceeded(bps, mongodb)

}
