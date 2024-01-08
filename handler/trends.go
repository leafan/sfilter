package handler

import (
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/global"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func HandleGlobalInfo(block *schema.Block, mongodb *mongo.Client) {
	// 先更新1min trend 趋势线
	updateGlobalInfo1MinTrends(block, mongodb)

	// 在更新config表中的全局字段(调用mongo统计即可)
	updateGlobalInfo(block, mongodb)
}

func updateGlobalInfo1MinTrends(block *schema.Block, mongodb *mongo.Client) {
	_time := block.BlockTime
	key := fmt.Sprintf("%v_%v_%v", _time.Day(), _time.Hour(), _time.Minute())

	// 先从db找到，如果没找到就new一个新的保存
	trends, err := global.GetTrendByKey(key, mongodb)
	if err == nil || (err == mongo.ErrNoDocuments) {
		// 全部赋值一遍即可
		trends.TimelineKey = key

		trends.TxNums += block.TxNums
		trends.PairCreatedNum += block.PairCreatedNum
		trends.VolumeByUsd += block.VolumeByUsd

		trends.BaseGas = block.Block.BaseFee().Int64()
		trends.Timestamp = _time

		global.UpsertTrends(trends, mongodb)
	}

}

// 直接用 mongo aggregate 功能即可统计..
// 不过由于有1h和24h的比较, 直接取出来go自己干可能对mongo更友好
func updateGlobalInfo(block *schema.Block, mongodb *mongo.Client) {
	info := &schema.GlobalInfo{
		ConfigKey: global.GlobalInfoKey,
	}

	// 每n个区块执行一次即可..
	if block.Block.Number().Int64()%config.GlobalUpdateIntervalBlocks != 0 {
		return
	}

	info.BaseGasPrice = block.Block.BaseFee().Int64()

	updateGlobalInfoFor1h(info, mongodb)
	updateGlobalInfoFor24h(info, mongodb)

	global.UpdateGlobalInfo(info, mongodb)
}

func updateGlobalInfoFor1h(info *schema.GlobalInfo, mongodb *mongo.Client) {
	end := time.Now()
	start := end.Add(-(60*2 + 1) * time.Minute) // 倒查2小时
	trends, err := global.GetTrendsByTimeRange(start, end, mongodb)

	// 一般情况下长度都是正常的
	length := len(trends)
	if err != nil || length < 108 {
		utils.Warnf("[ updateGlobalInfoFor1h ] err or length is to short. err: %v, len: %v\n", err, length)
		return
	}

	half := length / 2
	latest := trends[half:]
	last := trends[0:half]

	// 更新 TxNumIn1h, VolumeByUsdIn1h
	for _, trend := range latest {
		info.TxNumIn1h += trend.TxNums
		info.PairCreatedNumIn1h += trend.PairCreatedNum
		info.VolumeByUsdIn1h += trend.VolumeByUsd
	}

	var lastInfo = schema.GlobalInfo{}
	for _, trend := range last {
		lastInfo.TxNumIn1h += trend.TxNums
		lastInfo.PairCreatedNumIn1h += trend.PairCreatedNum
		lastInfo.VolumeByUsdIn1h += trend.VolumeByUsd
	}
	lastInfo.BaseGasPrice = last[len(last)-1].BaseGas // 上一个小时最新的一个

	info.TxNumChangeIn1h = utils.CalcChange(float64(info.TxNumIn1h), float64(lastInfo.TxNumIn1h))
	info.PairCreatedChangeIn1h = utils.CalcChange(float64(info.PairCreatedNumIn1h), float64(lastInfo.PairCreatedNumIn1h))
	info.BaseGasPriceChgIn1h = utils.CalcChange(float64(info.BaseGasPrice), float64(lastInfo.BaseGasPrice))
	info.VolumeChangeByUsdIn1h = utils.CalcChange(info.VolumeByUsdIn1h, lastInfo.VolumeByUsdIn1h)
}

func updateGlobalInfoFor24h(info *schema.GlobalInfo, mongodb *mongo.Client) {
	// 1min柱子, 2天有 60*24=2880根, 1根柱子大小 36B 以内, 共 100k不到
	end := time.Now()
	start := end.Add(-(60*24*2 + 1) * time.Minute) // 倒查48小时
	trends, err := global.GetTrendsByTimeRange(start, end, mongodb)

	// 一般情况下长度都是正常的, 如果不一样, 可能是某小时丢数据了或者卡块了
	// 允许一定的误差, 也就是丢失 60 个块数据
	length := len(trends)
	if err != nil || length < (60*24*2-60) {
		utils.Warnf("[ updateGlobalInfoFor24h ] err or length is to short. err: %v, len: %v\n", err, length)
		return
	}

	half := length / 2
	latest := trends[half:] //  保证最新长度
	last := trends[0:half]

	// 更新 TxNumIn1h, VolumeByUsdIn1h
	for _, trend := range latest {
		info.TxNumIn24h += trend.TxNums
		info.PairCreatedNumIn24h += trend.PairCreatedNum
		info.VolumeByUsdIn24h += trend.VolumeByUsd
	}

	var lastInfo = schema.GlobalInfo{}
	for _, trend := range last {
		lastInfo.TxNumIn24h += trend.TxNums
		lastInfo.PairCreatedNumIn24h += trend.PairCreatedNum
		lastInfo.VolumeByUsdIn24h += trend.VolumeByUsd
	}
	lastInfo.BaseGasPrice = last[len(last)-1].BaseGas

	info.TxNumChangeIn24h = utils.CalcChange(float64(info.TxNumIn24h), float64(lastInfo.TxNumIn24h))
	info.PairCreatedChangeIn24h = utils.CalcChange(float64(info.PairCreatedNumIn24h), float64(lastInfo.PairCreatedNumIn24h))
	info.BaseGasPriceChgIn24h = utils.CalcChange(float64(info.BaseGasPrice), float64(lastInfo.BaseGasPrice))
	info.VolumeChangeByUsdIn24h = utils.CalcChange(info.VolumeByUsdIn24h, lastInfo.VolumeByUsdIn24h)
}
