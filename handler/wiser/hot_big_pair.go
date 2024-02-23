package handler

/**
 * @20240222

pair条件(取与):
0. 非通缩币
1. pair创建一个月以上
2. 池子有效币值大于10万美金(v2为双边, v3为双边有效币种市值)
3. 最近24h的币价涨幅大于0

以及:
a —— 日线:
	1. 最近一天交易量超过最近半个月交易量均值30%以上
	2. 最新价格超过最近半个月均值5%以上
b —— 4小时线:
	1. 最近4小时交易量超过最近15根柱子的4小时线交易量均值30%以上
	2. 最新价格价格超过最近15根柱子4小时线的收盘价均值5%以上
c —— 1小时线:
	1. 最近1小时交易量超过最近15根1小时线交易量均值30%以上
	2. 最新价格价格超过最近15根1小时线的收盘价均值5%以上

以及:
最新柱子(如24h, 4h, 1h)的交易量为新高
*/

import (
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/kline"
	"sfilter/services/pair"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// hot pair
type HBPair struct {
	set *Setting
}

func (p *HBPair) Run() {
	interval := time.Duration(p.set.Config.HotPairCheckInterval) * time.Second
	timer := time.NewTicker(interval)
	defer timer.Stop()

	p.HotPairSearch()
	for range timer.C {
		p.HotPairSearch()
	}
}

func (p *HBPair) HotPairSearch() {
	var hpairs1d []*schema.Pair
	var hpairs4h []*schema.Pair
	var hpairs1h []*schema.Pair

	// 先取出符合基础条件的pair, 并按24h交易量增加量排序
	pairs := p.GetBasicPairs()

	for _, _pair := range pairs {
		if p.checkPairHot1Hour(_pair) {
			hpairs1h = append(hpairs1h, _pair)
		}

		if p.checkPairHot4Hour(_pair) {
			hpairs4h = append(hpairs4h, _pair)
		}

		if p.checkPairHot1Day(_pair) {
			hpairs1d = append(hpairs1d, _pair)
		}
	}

	p.printPairs(hpairs1h)
	p.printPairs(hpairs4h)
	p.printPairs(hpairs1d)
}

func (p *HBPair) printPairs(hpairs []*schema.Pair) {
	utils.Infof("[ printPairs ] hpairs len: %v", len(hpairs))

	for _, pair := range hpairs {
		fmt.Printf("%v, %v\n", pair.Address, pair.PairName)
	}
}

// 其他基础条件已满足, 需判断k线相关逻辑
func (p *HBPair) checkPairHot1Hour(_pair *schema.Pair) bool {
	now := time.Now()
	days := 1

	// 柱子是升序排列
	klines1Hour := kline.Get1HourKlineWithFullGenerated(_pair.Address, now, days, p.set.DB.Database(config.DatabaseName))

	if len(klines1Hour) < 1*24 {
		utils.Warnf("[ checkPairHot1Hour ] klines1Hour len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}

	// 去掉最新的一根柱子, 因为他可能还未跑完, 非整点
	// 同时只取最新15根柱子
	klines1Hour = klines1Hour[len(klines1Hour)-14 : len(klines1Hour)-1]

	volumeCur := _pair.VolumeByUsdIn1h
	priceCur := _pair.PriceInUsd
	if volumeCur == 0 || priceCur == 0 {
		utils.Warnf("[ checkPairHot1Hour ] current data is zero. return. pairName: %v, addr: %v, volumeCur: %v, priceCur: %v", _pair.PairName, _pair.Address, volumeCur, priceCur)
		return false
	}

	var volumeMid, priceMid float64
	for _, kline := range klines1Hour {
		volumeMid += kline.VolumeInUsd
		priceMid += kline.PriceInUsd
		// utils.Tracef("[ debug ], kline volume: %v, price: %v", kline.VolumeInUsd, kline.PriceInUsd)
	}

	// 取均值
	volumeMid = volumeMid / float64(len(klines1Hour))
	priceMid = priceMid / float64(len(klines1Hour))

	// 取增量
	volumeIncreament := (volumeCur - volumeMid) / volumeCur
	priceIncreament := (priceCur - priceMid) / priceCur

	// 条件比较
	if volumeIncreament < p.set.Config.VolumeIncrement {
		utils.Warnf("[ checkPairHot1Hour ] VolumeIncrement less, return. pairName: %v, addr: %v, volumeIncreament: %v", _pair.PairName, _pair.Address, volumeIncreament)
		return false
	}

	if priceIncreament < p.set.Config.MinPriceIncreament || priceIncreament > p.set.Config.MaxPriceIncreament {
		utils.Warnf("[ checkPairHot1Hour ] PriceIncreament less, return. pairName: %v, addr: %v, priceIncreament: %v, priceCur: %v, priceMid: %v", _pair.PairName, _pair.Address, priceIncreament, priceCur, priceMid)
		return false
	}

	// 需要最新柱子的volume是历史柱子中最大的
	for _, kline := range klines1Hour {
		if volumeCur < kline.VolumeInUsd {
			utils.Warnf("[ checkPairHot1Hour ] volume not the biggest, return. pairName: %v, addr: %v, volumeCur: %v, kline.VolumeInUsd: %v", _pair.PairName, _pair.Address, volumeCur, kline.VolumeInUsd)
			return false
		}
	}

	utils.Infof("[ checkPairHot1Hour ] pass. pairName: %v, addr: %v, volumeInc: %v, priceInc: %v", _pair.PairName, _pair.Address, volumeIncreament, priceIncreament)

	return true
}

func (p *HBPair) checkPairHot4Hour(_pair *schema.Pair) bool {
	now := time.Now()
	days := 3

	// 柱子是升序排列
	klines1Hour := kline.Get1HourKlineWithFullGenerated(_pair.Address, now, days, p.set.DB.Database(config.DatabaseName))

	if len(klines1Hour) < 3*24 {
		utils.Warnf("[ checkPairHot4Hour ] klines1Hour len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}

	// 去掉最新的一根柱子, 因为他可能还未跑完, 非整点
	klines1Hour = klines1Hour[len(klines1Hour)-11 : len(klines1Hour)-1]

	// 需要通过分钟线取最近4小时的交易量
	var volumeCur float64
	klines1Min := kline.Get1MinKlineWithFullGenerated(_pair.Address, now, 4, p.set.DB.Database(config.DatabaseName))
	for _, kl1m := range klines1Min {
		volumeCur += kl1m.VolumeInUsd
	}

	priceCur := _pair.PriceInUsd
	if volumeCur == 0 || priceCur == 0 {
		utils.Warnf("[ checkPairHot4Hour ] current data is zero. return. pairName: %v, addr: %v, volumeCur: %v, priceCur: %v", _pair.PairName, _pair.Address, volumeCur, priceCur)
		return false
	}

	// 每隔4小时取个值
	var klineNum int
	var volumeMid, priceMid float64
	for ind, kline := range klines1Hour {
		volumeMid += kline.VolumeInUsd
		if ind%4 == 0 {
			klineNum++
			priceMid += kline.PriceInUsd
			// utils.Tracef("[ debug ], kline volume: %v, price: %v", kline.VolumeInUsd, kline.PriceInUsd)
		}
	}

	// 取均值
	volumeMid = volumeMid / float64(klineNum)
	priceMid = priceMid / float64(klineNum)

	// 取增量
	volumeIncreament := (volumeCur - volumeMid) / volumeCur
	priceIncreament := (priceCur - priceMid) / priceCur

	// 条件比较
	if volumeIncreament < p.set.Config.VolumeIncrement {
		utils.Warnf("[ checkPairHot4Hour ] VolumeIncrement less, return. pairName: %v, addr: %v, volumeIncreament: %v", _pair.PairName, _pair.Address, volumeIncreament)
		return false
	}

	if priceIncreament < p.set.Config.MinPriceIncreament || priceIncreament > p.set.Config.MaxPriceIncreament {
		utils.Warnf("[ checkPairHot4Hour ] PriceIncreament less, return. pairName: %v, addr: %v, priceIncreament: %v, priceCur: %v, priceMid: %v", _pair.PairName, _pair.Address, priceIncreament, priceCur, priceMid)
		return false
	}

	// 需要最新柱子的volume是历史柱子中最大的
	var candiskVolume float64
	for ind, kline := range klines1Hour {
		candiskVolume += kline.VolumeInUsd
		if ind%4 == 0 {
			candiskVolume = 0
		}

		if volumeCur < candiskVolume {
			utils.Warnf("[ checkPairHot1Hour ] volume not the biggest, return. pairName: %v, addr: %v, volumeCur: %v, candiskVolume: %v", _pair.PairName, _pair.Address, volumeCur, candiskVolume)
			return false
		}
	}

	utils.Infof("[ checkPairHot4Hour ] pass. pairName: %v, addr: %v, volumeInc: %v, priceInc: %v", _pair.PairName, _pair.Address, volumeIncreament, priceIncreament)

	return true
}

func (p *HBPair) checkPairHot1Day(_pair *schema.Pair) bool {
	now := time.Now()
	days := 15

	// 柱子是升序排列
	klines1Hour := kline.Get1HourKlineWithFullGenerated(_pair.Address, now, days, p.set.DB.Database(config.DatabaseName))

	if len(klines1Hour) < 15*24 {
		utils.Errorf("[ checkPairHot1Hour ] klines1Hour len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}

	// 去掉最新的一根柱子, 因为他可能还未跑完, 非整点
	klines1Hour = klines1Hour[:len(klines1Hour)-1]

	volumeCur := _pair.VolumeByUsdIn24h
	priceCur := _pair.PriceInUsd
	if volumeCur == 0 || priceCur == 0 {
		utils.Warnf("[ checkPairHot1Hour ] current data is zero. return. pairName: %v, addr: %v, volumeCur: %v, priceCur: %v", _pair.PairName, _pair.Address, volumeCur, priceCur)
		return false
	}

	// 每隔24小时取个值
	var klineNum int
	var volumeMid, priceMid float64
	for ind, kline := range klines1Hour {
		volumeMid += kline.VolumeInUsd // volume需要一直加

		if ind%24 == 0 {
			klineNum++
			priceMid += kline.PriceInUsd
			// utils.Tracef("[ debug ], kline volume: %v, price: %v", kline.VolumeInUsd, kline.PriceInUsd)
		}
	}

	// 取均值
	volumeMid = volumeMid / float64(klineNum)
	priceMid = priceMid / float64(klineNum)

	// 取增量
	volumeIncreament := (volumeCur - volumeMid) / volumeCur
	priceIncreament := (priceCur - priceMid) / priceCur

	// 条件比较
	if volumeIncreament < p.set.Config.VolumeIncrement {
		utils.Warnf("[ checkPairHot1Hour ] VolumeIncrement less, return. pairName: %v, addr: %v, volumeIncreament: %v", _pair.PairName, _pair.Address, volumeIncreament)
		return false
	}

	if priceIncreament < p.set.Config.MinPriceIncreament || priceIncreament > p.set.Config.MaxPriceIncreament {
		utils.Warnf("[ checkPairHot1Hour ] PriceIncreament less, return. pairName: %v, addr: %v, priceIncreament: %v, priceCur: %v, priceMid: %v", _pair.PairName, _pair.Address, priceIncreament, priceCur, priceMid)
		return false
	}

	// 需要最新柱子的volume是历史柱子中最大的
	var candiskVolume float64
	for ind, kline := range klines1Hour {
		candiskVolume += kline.VolumeInUsd
		if ind%24 == 0 {
			candiskVolume = 0
		}

		if volumeCur < candiskVolume {
			utils.Warnf("[ checkPairHot1Hour ] volume not the biggest, return. pairName: %v, addr: %v, volumeCur: %v, candiskVolume: %v", _pair.PairName, _pair.Address, volumeCur, candiskVolume)
			return false
		}
	}

	utils.Infof("[ checkPairHot1Hour ] pass. pairName: %v, addr: %v, volumeInc: %v, priceInc: %v", _pair.PairName, _pair.Address, volumeIncreament, priceIncreament)

	return true
}

func (p *HBPair) GetBasicPairs() []*schema.Pair {
	// test
	// pairT, _ := pair.GetPairInfoForRead("0x769f539486b31eF310125C44d7F405C6d470cD1f")
	// var pairs []*schema.Pair
	// pairs = append(pairs, pairT)
	// return pairs

	filter := bson.M{}

	// 池子 tvl 要求
	filter["liquidityInUsd"] = bson.M{
		"$gte": p.set.Config.MinPairLiquidity,
	}

	// 价格最小涨幅要求, 减少数据量
	filter["priceChangeIn24h"] = bson.M{
		"$gte": 0,
	}

	// pair创建时长要求
	date := time.Now().Add(-time.Duration(p.set.Config.MinPairCreatAge) * time.Second)
	filter["firstAddPoolTime"] = bson.M{
		"$lte": date,
	}

	// 基础排序策略
	limit := int64(1000)
	options := &options.FindOptions{Limit: &limit}

	sort := bson.D{bson.E{Key: "txNumIn24h", Value: -1}}
	options = options.SetSort(sort)

	db := p.set.DB.Database(config.DatabaseName)
	pairs, _, err := pair.GetHotPairs(options, &filter, db)
	if err != nil {
		utils.Errorf("[ GetBasicPairs ] GetHotPairs err: %v", err)
		return nil
	}

	return pairs
}
