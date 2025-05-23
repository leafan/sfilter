package handler

import (
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	userModels "sfilter/user/models"
	"sfilter/utils"

	"sfilter/services/swap"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// 以地址做key, value是所有监控了这个地址的用户列表
var trackAddressMap userModels.UserTrackAddressMap

// 近期有更新地址的user, 用于判断是否需要检查记录条数上限
// 弄这个全局变量的主要目的是为了性能考虑, 只有近期有更新的才检查
var trackUserMap = make(map[string]bool)

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
	utils.Infof("[ GetTrackAddressMapOnTimer ] time consumed: %v, unique address count: %v, total trackAddressMap kv count: %v\n", time.Since(start), len(trackAddressMap), totalCount)

	return nil
}

// 检查用户的记录条数是否超限 1.2 倍, 如果是, 删除 30%
func CheckAndDeleteUserSwapsUpLimit(mongodb *mongo.Client) {
	for username := range trackUserMap {
		count, err := swap.GetUserTrackSwapCount(username, mongodb)
		if err != nil {
			continue
		}

		user, err := userModels.GetUser(username)
		if err != nil {
			continue
		}

		upper := userModels.GetRoleTrackSwapCount(fmt.Sprintf("%d", user.Role))
		if count >= int64(float64(upper)*1.2) { // 超过1.2倍
			utils.Warnf("[ CheckAndDeleteUserSwapsUpLimit ] reach limit! delete.. count: %v, upper: %v", count, upper)

			// 批量删除最老的一定量数据
			deleteCount := count - int64(float64(upper)*0.8) // 删除至剩余80%
			if deleteCount > config.COUNT_UPPER_SIZE {
				deleteCount = config.COUNT_UPPER_SIZE // 太大了, 下次再来吧..
			}

			swap.DeleteOldEntries(username, deleteCount, mongodb)
		}
	}

}

// 将用户匹配出的swap信息逐一写入db
// 这里要注意某些用户的容量并及时告警
func HandleUserTrackSwaps(block *schema.Block, mongodb *mongo.Client, swaps []*schema.Swap) {
	if len(trackAddressMap) <= 0 {
		utils.Warnf("[ HandleUserTrackSwaps ] len trackAddressMap is zero? no init?")
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

				TxHash:         _swap.TxHash,
				LogIndexWithTx: _swap.LogIndexWithTx,

				Price:             _swap.Price,
				AmountOfMainToken: _swap.AmountOfMainToken,
				VolumeInUsd:       _swap.VolumeInUsd,

				SwapTime: _swap.SwapTime,
			},
		}

		swap.SaveTrackSwap(uswap, mongodb)

		// 记录该用户状态, 用于检查是否达到存储上限
		trackUserMap[uswap.Username] = true

		// utils.Tracef("[ doHandleOneSwap ] save one now, user: %v", uswap.Username)
	}

}
