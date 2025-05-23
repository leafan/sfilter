package handler

import (
	"context"
	"math/big"
	"sync"
	"time"

	"sfilter/config"
	"sfilter/schema"
	service_block "sfilter/services/block"
	"sfilter/services/chain"
	"sfilter/utils"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

// 最终需要将block、transfer等都定义为类

type Handler struct {
	DB     *mongo.Client
	Client *ethclient.Client

	Tokens  schema.TokenMap
	Pairs   schema.PairMap
	Routers schema.RouterMap

	SAddresses schema.SpecialAddressMap // 特殊地址如黑洞地址等

	SwapContracts map[string]bool // 把swap相关的地址全部存到一个地址, 方便查询
}

func NewHandler(client *ethclient.Client, db *mongo.Client) (*Handler, error) {
	h := &Handler{
		Client: client,
		DB:     db,
	}

	err := h.initMaps()
	return h, err
}

func (h *Handler) Run(block int64) {
	if block > 0 {
		h.debugBlock(block)
		return
	}

	go h.Retrive_old_blocks() // 先回溯

	headers := make(chan *types.Header)
	sub, err := h.Client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		utils.Fatalf("SubscribeNewHead error: %v", err)
	}
	utils.Infof("[ loop ] start SubscribeNewHead now..\n\n")

	for {
		select {
		case err := <-sub.Err():
			utils.Warnf("[ Run ]SubscribeBlocks error: %v, reconnecting now...", err)

			for {
				sub, err = h.Client.SubscribeNewHead(context.Background(), headers)
				if err != nil {
					utils.Warnf("[ Run ] SubscribeNewHead err: %v, sleep now..", err)
					time.Sleep(5 * time.Second)
				} else {
					utils.Infof("[ Run ] reconnect success..")
					break
				}

			}

		case header := <-headers:
			utils.Infof("[ loop ] Get new header now. number: %v\n", header.Number)
			go h.HandleBlock(header.Number)
		}

	}
}

func (h *Handler) HandleBlock(blockNumber *big.Int) {
	block, err := getBlock(blockNumber, h.Client, h.DB, 0)

	if err == nil {
		h.handleOneBlock(block)
	}
}

// 每次启动往回回溯n个区块, 防止某一次未处理
// 回溯的时候, eth价格通过infura获取, 每10个区块更新一次价格
func (h *Handler) Retrive_old_blocks() {
	if config.RetriveOldBlockNum < 0 {
		utils.Infof("[ Retrive_old_blocks ] no retrive, return..")
		return
	}

	curBlkNo, err := h.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		utils.Fatalf("[ Retrive_old_blocks ] HeaderByNumber err: %v", err)
	}

	startBlock := curBlkNo.Number.Int64() - int64(config.RetriveOldBlockNum)
	utils.Infof("[ Retrive_old_blocks ] retrive now.. start block: %v", startBlock)

	var ethPrice float64
	times := 0

	// 控制并发执行的协程数量
	ch := make(chan struct{}, config.MaxConcurrentRoutineNums)
	var wg sync.WaitGroup

	for i := startBlock; i < curBlkNo.Number.Int64(); i++ {
		if service_block.IsBlockProceeded(i, h.DB) {
			utils.Debugf("[ Retrive_old_blocks ] block %v has handled, pass..", i)
			continue
		}

		if times%config.GetPriceIntervalForRetrive == 0 {
			ethPrice, err = chain.GetBasicCoinPrice(h.Client, big.NewInt(i), config.BlockChain)
			if err != nil {
				utils.Warnf("[ Retrive_old_blocks ] GetBasicCoinPrice err: %v", err)
				continue // eth价格必须取到, 如果没取到, 回溯
			}
		}

		ch <- struct{}{} // 当协程处理不过来的时候, 这里会阻塞，写不进去
		wg.Add(1)

		go func(i int64, ethPrice float64) {
			defer wg.Done()

			block, err := getBlock(big.NewInt(i), h.Client, h.DB, ethPrice)
			if err != nil {
				return
			}

			<-ch // 处理完了, 释放一个协程位置

			// 先释放协程, 再往后走, 因为这里面有 sleep
			if block != nil {
				h.handleOneBlock(block)
			}

		}(i, ethPrice)

		times++

		time.Sleep(config.SleepIntervalforRetrive * time.Millisecond)
	}

	close(ch)
	wg.Wait()
}

func (h *Handler) debugBlock(block int64) {
	config.DevelopmentMode = true

	// 先把block id set 未处理
	service_block.SetUnProceeded(block, h.DB)
	h.HandleBlock(big.NewInt(block))
}
