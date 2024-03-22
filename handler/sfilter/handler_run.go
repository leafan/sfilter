package handler

import (
	"sfilter/config"
	"sfilter/schema"
	service_block "sfilter/services/block"
	"sfilter/services/pair"
	"sfilter/services/token"
	"sfilter/utils"
	"sync"
	"time"
)

var handler_Lock sync.Mutex

// 更新 pair map和token map等, 防止程序未重新拉取db导致 数据没有及时更新
func (h *Handler) updateMapBySwaps(swaps []*schema.Swap) {
	handler_Lock.Lock()
	defer handler_Lock.Unlock()

	for _, _swap := range swaps {
		// 更新pair map
		_, ok := h.Pairs[_swap.PairAddr]
		if !ok {
			_pair, err := pair.GetPairInfoForRead(_swap.PairAddr)
			if err == nil {
				h.Pairs[_swap.PairAddr] = _pair
				h.SwapContracts[_swap.PairAddr] = true
			}
		}

		// 更新token map
		_, ok = h.Tokens[_swap.MainToken]
		if !ok {
			_token, err := token.GetTokenInfo(_swap.MainToken, h.DB)
			if err == nil {
				h.Tokens[_swap.MainToken] = _token
				h.SwapContracts[_swap.MainToken] = true
			}
		}
	}
}

// 串行处理即可, 因为跑到这里来的都是协程
func (h *Handler) handleOneBlock(blk *schema.Block) {
	start := time.Now()

	HandlePairLogic(blk, h.DB)
	HandleLiquidityLogic(blk, h.DB)

	transfers := HandleTransfer(blk, h.DB) // 获取transfer信息

	swaps := HandleSwapAndKline(blk, h.DB) // 获取swaps
	// 针对每一笔swap, 把里面碰到的 pair, token 都更新一下, 防止不及时
	h.updateMapBySwaps(swaps)

	// 更新swap的trader等信息, 然后写入db
	UpsertSwapToDB(swaps, h.SwapContracts, transfers, h.DB)

	// 更新transfer的usd value等信息, 然后写入db
	UpsertTransferToDB(transfers, swaps, h.DB)

	// trade info 是更新最近24h或7天的数据, 因此老数据就别掺和了
	if time.Since(time.Unix(int64(blk.Block.Time()), 0)).Seconds() < config.SecondsForOneWeek {
		HandleTradeInfo(blk, h.DB, swaps)
		HandleGlobalInfo(blk, h.DB)
	}

	// 更新 用户跟踪地址逻辑
	// HandleUserTrackSwaps(blk, h.DB, swaps)

	// 更新 facet 逻辑
	// facet.HandleFacetLogic(blk, h.DB)

	// etc.. todo

	// record the proceeded block.
	h.setBlockToProceeded(blk)

	utils.Debugf("Handle block: %d finished, swap num: %v, time elapsed: % v\n", blk.Block.NumberU64(), blk.TxNums, time.Since(start))
}

func (h *Handler) initMaps() error {
	var err error
	dbName := config.DatabaseName

	h.Tokens, err = token.GetTokenMap(config.SELECT_UPPER_SIZE, h.DB.Database(dbName))
	if err != nil {
		utils.Fatalf("[ InitMap ] GetTokenMap failed: %v", err)
	}

	h.Pairs, err = pair.GetPairMap(config.SELECT_UPPER_SIZE, h.DB.Database(dbName))
	if err != nil {
		utils.Fatalf("[ InitMap ] GetPairMap failed: %v", err)
	}

	// 继续 routers 和 saddresss等
	h.SAddresses = make(schema.SpecialAddressMap)
	h.Routers = make(schema.RouterMap)

	for _, key := range config.BlackHoleAddresses {
		h.SAddresses[key] = schema.SpecialAddress{Address: key}
	}

	for _, key := range config.FamousRouters {
		h.Routers[key] = schema.Router{Address: key}
	}

	// 把需要和swap判断相关的地址存到一个map里, 方便查询
	h.SwapContracts = make(map[string]bool)
	for key := range h.Pairs {
		h.SwapContracts[key] = true
	}
	for key := range h.Tokens {
		h.SwapContracts[key] = true
	}
	for key := range h.Routers {
		h.SwapContracts[key] = true
	}
	for key := range h.SAddresses {
		h.SwapContracts[key] = true
	}

	utils.Infof("[ InitMap ] tokens len: %v, pairs len: %v", len(h.Tokens), len(h.Pairs))
	return nil
}

func (h *Handler) setBlockToProceeded(block *schema.Block) {
	bps := &schema.BlockProceeded{
		BlockNo:     block.Block.Number().Int64(),
		Hash:        block.Block.Hash().String(),
		BlockTime:   int64(block.Block.Time()),
		TxNums:      block.TxNums,
		VolumeByUsd: block.VolumeByUsd,

		EthPrice: block.EthPrice,
	}

	service_block.SetBlockProceeded(bps, h.DB)

}
