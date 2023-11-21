package handler

import (
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/kline"
	services_pair "sfilter/services/pair"
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
	pair, err := chain.GetPairInfo(_pair, mongodb)
	if err != nil {
		log.Printf("[ updatePairInfo ] GetTokenInfo wrong, return.. err: %v\n\n", err)
		return
	}

	// 为避免mongo反复读取, 使用kline相关信息, 然后进行计算更新
	// 只需要2次get操作(分钟线与小时线), 就可以全部计算完毕

	// 取最近2-3小时数据, 可以计算出2h内的数据变化(Change)
	now := time.Now()
	klines1Min := kline.Get1MinKlineWithFullGenerated(pair.Address, now, 2, mongodb)
	klines1Hour := kline.Get1HourKlineWithFullGenerated(pair.Address, now, 2, mongodb)
	// log.Printf("\n\n\n[ updatePairInfo ] pair: %v, time is: %v, klines1Min: %v\n\n\n\n", pair.Address, now, klines1Min)

	if len(klines1Min) != 120 { // 120根柱子
		log.Printf("[updatePairInfo] kline 1min is empty, return.. pair: %v, len: %v\n", _pair, len(klines1Min))
		return
	}
	updatePairTx1h(klines1Min, pair)
	updatePairPrice1h(klines1Min, pair)

	if len(klines1Hour) == 48 {
		updatePairTx24h(klines1Hour, pair)
		updatePairPrice24h(klines1Hour, pair)
	} else {
		log.Printf("[updatePairInfo] kline 1hour is empty, return.. pair: %v, len: %v\n", _pair, len(klines1Hour))
	}

	// update to db
	services_pair.UpSertPairInfo(pair, mongodb)

	log.Printf("\n\n\n[ updatePairInfo ] update finished... pair: %v\n", _pair)
}

func updatePairTx1h(klines1Min []schema.KLine, pair *schema.Pair) {
	// 先更新 TxNumIn1h
	var txNum1h, txNumBefore1h int

	for i := 120 - 1; i >= 60; i-- {
		txNum1h += klines1Min[i].TxNum
	}
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

	pair.TxNumIn1h = txNum1h
	pair.TxNumChangeIn1h = change
}

func updatePairTx24h(klines1Hour []schema.KLine, pair *schema.Pair) {
	var txNum24h, txNumBefore24h int

	for i := 48 - 1; i >= 24; i-- {
		txNum24h += klines1Hour[i].TxNum
	}
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

	pair.TxNumIn24h = txNum24h
	pair.TxNumChangeIn24h = change
}

func updatePairPrice1h(klines1Min []schema.KLine, pair *schema.Pair) {
	var volume float64
	var change float32

	pair.Price = klines1Min[len(klines1Min)-1].ClosePrice

	current := klines1Min[len(klines1Min)-1].ClosePrice
	last := klines1Min[len(klines1Min)-60-1].ClosePrice
	if last != 0 {
		change = float32((current - last) / last)
	}
	pair.PriceChangeIn1h = change

	for i := 120 - 1; i >= 60; i-- {
		volume += klines1Min[i].VolumeInUsd
	}
	pair.VolumeByUsdIn1h = volume
}

func updatePairPrice24h(klines1Hour []schema.KLine, pair *schema.Pair) {
	var volume float64
	var change float32

	current := klines1Hour[len(klines1Hour)-1].ClosePrice
	last := klines1Hour[len(klines1Hour)-24-1].ClosePrice
	if last != 0 {
		change = float32((current - last) / last)
	}
	pair.PriceChangeIn24h = change

	for i := 48 - 1; i >= 24; i-- {
		volume += klines1Hour[i].VolumeInUsd
	}
	pair.VolumeByUsdIn24h = volume
}
