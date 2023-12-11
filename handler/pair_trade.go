package handler

import (
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/kline"
	services_pair "sfilter/services/pair"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// trade info 放到pair 中，方便直接查询读取等
func HandleTradeInfo(block *schema.Block, mongodb *mongo.Client, swaps []*schema.Swap) {
	// 先取出本次需要更新的token信息
	pairs := make(map[string]int)

	for _, swap := range swaps {
		pairs[swap.PairAddr]++
	}

	for key := range pairs {
		updatePairInfo(key, mongodb)
	}

}

func updatePairInfo(_pair string, mongodb *mongo.Client) {
	pair, err := services_pair.GetPairInfo(_pair, mongodb)
	if err != nil {
		utils.Errorf("[ updatePairInfo ] GetTokenInfo wrong, return.. err: %v\n\n", err)
		return
	}

	// 为避免mongo反复读取, 使用kline相关信息, 然后进行计算更新
	// 只需要2次get操作(分钟线与小时线), 就可以全部计算完毕

	// 取最近2-3小时数据, 可以计算出2h内的数据变化(Change)
	now := time.Now()
	klines1Min := kline.Get1MinKlineWithFullGenerated(pair.Address, now, 2, mongodb)
	klines1Hour := kline.Get1HourKlineWithFullGenerated(pair.Address, now, 2, mongodb)

	updatePairTx1h(klines1Min, pair)
	updatePairPrice1h(klines1Min, pair)

	updatePairTx24h(klines1Hour, pair)
	updatePairPrice24h(klines1Hour, pair)

	// 纠正, 针对某些刚开始时候的pair
	if pair.TxNumIn24h < pair.TxNumIn1h {
		pair.TxNumIn24h = pair.TxNumIn1h
	}
	if pair.VolumeByUsdIn24h < pair.VolumeByUsdIn1h {
		pair.VolumeByUsdIn24h = pair.VolumeByUsdIn1h
	}

	// update to db
	services_pair.UpdateTradeInfo(pair, mongodb)
}

func updatePairTx1h(klines1Min []schema.KLine, pair *schema.Pair) {
	if len(klines1Min) <= 0 {
		return
	}

	var txNum1h int
	for i, j := len(klines1Min)-1, 0; i >= 0 && j < 60; i, j = i-1, j+1 {
		txNum1h += klines1Min[i].TxNum
	}
	pair.TxNumIn1h = txNum1h

	if len(klines1Min) != 120 {
		return
	}

	var txNumBefore1h int
	for i := 60 - 1; i >= 0; i-- {
		txNumBefore1h += klines1Min[i].TxNum
	}

	delta := txNum1h - txNumBefore1h
	var change float32

	if delta == 0 {
		change = 0
	} else if txNumBefore1h == 0 {
		change = config.INFINITE_CHANGE // 说明txNum1h一定大于0
	} else {
		change = float32(delta / txNumBefore1h)
	}

	pair.TxNumChangeIn1h = change
}

func updatePairTx24h(klines1Hour []schema.KLine, pair *schema.Pair) {
	if len(klines1Hour) <= 0 {
		return
	}

	var txNum24h int
	for i, j := len(klines1Hour)-1, 0; i >= 0 && j < 24; i, j = i-1, j+1 {
		txNum24h += klines1Hour[i].TxNum
	}
	pair.TxNumIn24h = txNum24h

	if len(klines1Hour) != 48 {
		return
	}

	var txNumBefore24h int
	for i := 24 - 1; i >= 0; i-- {
		txNumBefore24h += klines1Hour[i].TxNum
	}

	delta := txNum24h - txNumBefore24h
	var change float32

	if delta == 0 {
		change = 0
	} else if txNumBefore24h == 0 {
		change = config.INFINITE_CHANGE // 说明txNum1h一定大于0
	} else {
		change = float32(delta / txNumBefore24h)
	}
	pair.TxNumChangeIn24h = change
}

func updatePairPrice1h(klines1Min []schema.KLine, pair *schema.Pair) {
	if len(klines1Min) <= 0 {
		return
	}

	var volume float64
	for i, j := len(klines1Min)-1, 0; i >= 0 && j < 60; i, j = i-1, j+1 {
		volume += klines1Min[i].VolumeInUsd
	}
	pair.VolumeByUsdIn1h = volume
	pair.Price = klines1Min[len(klines1Min)-1].ClosePrice

	if len(klines1Min) < 61 { // 前一小时最后一分钟即可
		return
	}

	var change float32

	current := klines1Min[len(klines1Min)-1].ClosePrice
	last := klines1Min[len(klines1Min)-60-1].ClosePrice
	if last != 0 {
		change = float32((current - last) / last)
	}
	pair.PriceChangeIn1h = change
}

func updatePairPrice24h(klines1Hour []schema.KLine, pair *schema.Pair) {
	if len(klines1Hour) <= 0 {
		return
	}

	var volume float64
	for i, j := len(klines1Hour)-1, 0; i >= 0 && j < 24; i, j = i-1, j+1 {
		volume += klines1Hour[i].VolumeInUsd
	}
	pair.VolumeByUsdIn24h = volume

	if len(klines1Hour) < 25 { // 前一天最后一小时即可
		return
	}

	var change float32
	current := klines1Hour[len(klines1Hour)-1].ClosePrice

	last := klines1Hour[len(klines1Hour)-24-1].ClosePrice
	if last != 0 {
		change = float32((current - last) / last)
	}
	pair.PriceChangeIn24h = change
}
