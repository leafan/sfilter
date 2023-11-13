package handler

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"sfilter/config"
	"sfilter/services/block"
	"sfilter/services/chain"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleBlock(blockNumber *big.Int, client *ethclient.Client, mongodb *mongo.Client) {
	block.HandleBlock(blockNumber, client, mongodb, 0)
}

// 每次启动往回回溯n个区块, 防止某一次未处理
// 回溯的时候, eth价格通过infura获取, 每10个区块更新一次价格
func Retrive_old_blocks(client *ethclient.Client, mongodb *mongo.Client) {
	curBlkNo, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal("[ retrive_old_blocks ] HeaderByNumber err: ", err)
	}

	startBlock := curBlkNo.Number.Int64() - int64(config.RetriveOldBlockNum)

	ethPrice := chain.GetEthPrice(client, big.NewInt(startBlock))
	for i := startBlock; i < curBlkNo.Number.Int64(); i++ {
		if i%10 == 0 {
			fmt.Println("Retrive_old_blocks debug i: ", i)
			ethPrice = chain.GetEthPrice(client, big.NewInt(i))
		}

		go block.HandleBlock(big.NewInt((i)), client, mongodb, ethPrice)
		time.Sleep(20 * time.Millisecond)
	}

}
