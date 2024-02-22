package handler

/**
 * @20240222

pair条件（取与）：
1. pair创建一个月以上
2. 池子有效币值大于10万美金(v2为双边, v3为双边有效币种市值)
3. 最近一天交易量超过最近半个月交易量均值30%以上
4. 最近一天价格超过最近半个月均值5%以上
5. 最近24h的币价涨幅大于0
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
type HPair struct {
	set *Setting
}

func (p *HPair) Run() {
	interval := time.Duration(p.set.Config.HotPairCheckInterval) * time.Second
	timer := time.NewTicker(interval)
	defer timer.Stop()

	p.HotPairSearch()
	for range timer.C {
		p.HotPairSearch()
	}
}

func (p *HPair) HotPairSearch() {
	var hpairs []*schema.Pair

	// 先取出符合基础条件的pair, 并按24h交易量增加量排序
	pairs := p.GetBasicPairs()

	for _, _pair := range pairs {
		if p.checkPairHot(_pair) {
			hpairs = append(hpairs, _pair)
		}
	}

	p.printPairs(hpairs)
}

func (p *HPair) printPairs(hpairs []*schema.Pair) {
	utils.Infof("[ printPairs ] hpairs len: %v", len(hpairs))

	for _, pair := range hpairs {
		fmt.Printf("%v, %v\n", pair.Address, pair.PairName)
	}
}

// 其他基础条件已满足, 需判断k线相关逻辑
// 最近一天交易量超过最近半个月交易量均值30%以上, 价格超过最近半个月均值5%以上
func (p *HPair) checkPairHot(_pair *schema.Pair) bool {
	now := time.Now()
	days := 15

	// 柱子是升序排列
	klines1Hour := kline.Get1HourKlineWithFullGenerated(_pair.Address, now, days, p.set.DB.Database(config.DatabaseName))

	if len(klines1Hour) < 15*24 {
		utils.Errorf("[ checkPairHot ] klines1Hour len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}

	// 去掉最新的一根柱子, 因为他可能还未跑完, 非整点
	klines1Hour = klines1Hour[:len(klines1Hour)-1]

	volumeCur := klines1Hour[len(klines1Hour)-1].VolumeInUsd
	priceCur := klines1Hour[len(klines1Hour)-1].PriceInUsd
	if volumeCur == 0 || priceCur == 0 {
		utils.Warnf("[ checkPairHot ] current data is zero. return. pairName: %v, addr: %v, volumeCur: %v, priceCur: %v", _pair.PairName, _pair.Address, volumeCur, priceCur)
		return false
	}

	// 再去掉最新柱子, 因为那是最新值
	klines1Hour = klines1Hour[:len(klines1Hour)-1]

	var volumeMid, priceMid float64
	for _, kline := range klines1Hour {
		volumeMid += kline.VolumeInUsd
		priceMid += kline.PriceInUsd
	}

	// 取均值
	volumeMid = volumeMid / float64(len(klines1Hour))
	priceMid = priceMid / float64(len(klines1Hour))

	// 取增量
	volumeIncreament := (volumeCur - volumeMid) / volumeCur
	priceIncreament := (priceCur - priceMid) / priceCur

	// 条件比较
	if volumeIncreament < p.set.Config.VolumeIncrement {
		utils.Warnf("[ checkPairHot ] VolumeIncrement less, return. pairName: %v, addr: %v, volumeIncreament: %v", _pair.PairName, _pair.Address, volumeIncreament)
		// return false
	}

	if priceIncreament < p.set.Config.PriceIncreament {
		utils.Warnf("[ checkPairHot ] PriceIncreament less, return. pairName: %v, addr: %v, priceIncreament: %v, priceCur: %v, priceMid: %v", _pair.PairName, _pair.Address, priceIncreament, priceCur, priceMid)
		return false
	}

	utils.Infof("[ checkPairHot ] pass. pairName: %v, addr: %v", _pair.PairName, _pair.Address)

	return true
}

func (p *HPair) GetBasicPairs() []*schema.Pair {
	// test
	// pairT, _ := pair.GetPairInfoForRead("0x3416cF6C708Da44DB2624D63ea0AAef7113527C6")
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
