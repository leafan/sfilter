package handler

import (
	"sfilter/schema"
	userModels "sfilter/user/models"
	"sfilter/utils"
	gutils "sfilter/utils"

	"sfilter/services/swap"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var trackAddressMap userModels.UserTrackAddressMap

// 定时从db中获取address并更新
// 调用前已经初始化db
func GetTrackAddressMapOnTimer() error {
	var err error
	var totalCount int

	start := time.Now()
	trackAddressMap, totalCount, err = userModels.AdminGetTrackAddressMap()
	if err != nil {
		return err
	}
	gutils.Infof("[ GetTrackAddressMapOnTimer ] time consumed: %v, unique address count: %v, total trackAddressMap kv count: %v\n", time.Since(start), len(trackAddressMap), totalCount)

	return nil
}

// 将用户匹配出的swap信息逐一写入db
// 这里要注意某些用户的容量并及时告警
func HandleUserTrackSwaps(block *schema.Block, mongodb *mongo.Client, swaps []*schema.Swap) {
	if len(trackAddressMap) <= 0 {
		utils.Errorf("[ HandleUserTrackSwaps ] len trackAddressMap is zero? no init?")
		return
	}

	for i := 0; i < len(swaps); i++ {
		doHandleOneSwap(swaps[i], mongodb)
	}

}

// 处理某一个用户的swaps校验、保存等
func doHandleOneSwap(_swap *schema.Swap, mongodb *mongo.Client) {
	_, traderOk := trackAddressMap[_swap.Trader]
	_, operatorOK := trackAddressMap[_swap.Operator]

	if !traderOk && !operatorOK {
		return
	}

	tAddrs := trackAddressMap[_swap.Trader]
	if !traderOk {
		tAddrs = trackAddressMap[_swap.Operator]
	}

	// 写数据保存
	for i := 0; i < len(tAddrs); i++ {
		uswap := &schema.TrackSwap{
			UserAddrInfo: schema.UserAddrInfo{
				Username: tAddrs[i].Username,
				Address:  tAddrs[i].Address,
				Memo:     tAddrs[i].Memo,
				Priority: tAddrs[i].Priority,
			},
			TrackSwapInfo: schema.TrackSwapInfo{ // 每次都是一样信息, 空间换时间...
				PairAddr:  _swap.PairAddr,
				PairName:  _swap.PairName,
				MainToken: _swap.MainToken,
				Direction: _swap.Direction,

				TxHash: _swap.TxHash,

				Price:             _swap.Price,
				AmountOfMainToken: _swap.AmountOfMainToken,
				VolumeInUsd:       _swap.VolumeInUsd,

				SwapTime: _swap.SwapTime,
			},
		}

		swap.SaveTrackSwap(uswap, mongodb)
	}

}
