package handler

/**
 * @20240222

pair条件(取与):
0. 非通缩币
1. pair创建一个月以上
2. 池子有效币值大于10万美金(v2为双边, v3为双边有效币种市值)
3. 最近24h的币价涨幅大于0

4. 以及:
a —— 日线:
	1. 最近一天交易量超过最近半个月交易量均值x%以上
	2. 最新价格超过最近半个月均值5%以上
b —— 4小时线:
	1. 最近4小时交易量超过最近15根柱子的4小时线交易量均值x%以上
	2. 最新价格价格超过最近15根柱子4小时线的收盘价均值5%以上
c —— 1小时线:
	1. 最近1小时交易量超过最近15根1小时线交易量均值x%以上
	2. 最新价格价格超过最近15根1小时线的收盘价均值5%以上

@20240224 以及:
5. 最新柱子(如24h, 4h, 1h)的交易量为最高
6. 最新价格与过去15根柱子收盘价格差距不大于50%
7. 历史柱子中不存在0交易量

卖点:
亏50%或回调30%，回调定义：涨破成本50%后的下跌

*/

import (
	"fmt"
	"math"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/kline"
	"sfilter/services/pair"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/robfig/cron"
)

// hot pair
type HBPair struct {
	Set       *Setting
	traceTime time.Time

	hpairs1d []*schema.Pair
	hpairs4h []*schema.Pair
	hpairs1h []*schema.Pair
}

func (p *HBPair) init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	time.Local = loc

	// 将时间设置为 2024.02.23 0:0:00
	p.traceTime = time.Date(2024, 2, 22, 0, 0, 0, 1, loc)
	// p.traceTime = time.Now()
}

func (p *HBPair) Run() {
	p.init()

	{ // test
		p.PolicyLogic()
		return
	}

	c := cron.New()
	spec := "0 0 16 * * *" // 定时每晚执行, 注意机器是utc+0时区
	c.AddFunc(spec, func() {
		p.PolicyLogic()
	})
	c.Start()

	select {}
}

func (p *HBPair) PolicyLogic() {
	// 找出符合当前买点的pair
	p.HotPairSearch()

	//针对符合的pair找出卖点, 并回测数据
	// p.HotPairSell(p.hpairs1h)
	// p.HotPairSell(p.hpairs4h)
	p.HotPairSell(p.hpairs1d)
}

func (p *HBPair) HotPairSell(_pairs []*schema.Pair) {
	// p.printPairs(_pairs)

	utils.Infof("[ HotPairSell ] buy time: %v", p.traceTime.Format("2006-01-02 15:04:05"))

	for _, _pair := range _pairs {
		p.TraceBackOnePair(_pair)
	}
}

// 卖点: 亏50%或回调30%
// 回调定义: 涨破成本50%后的下跌
func (p *HBPair) TraceBackOnePair(_pair *schema.Pair) {
	curPrice := _pair.ReservedFloat

	// 取出从起始时间点到现在的分钟线, 取出价格
	// 由于高开低收取的是 pair 相对价格, 不是usd, 因此直接取收盘价usd
	now := time.Now()
	durationHours := now.Sub(p.traceTime).Hours()
	klines1Min := kline.Get1MinKlineWithFullGenerated(_pair.Address, now, int(durationHours), p.Set.DB.Database(config.DatabaseName))
	// utils.Debugf("[ TraceBackOnePair ] len for klines1Min: %v, curPrice: %v", len(klines1Min), curPrice)

	// 时间从过去到现在排序
	var isRetrace bool  // 是否属于回调阶段
	var highest float64 // 回溯期间最高价格
	for _, k1m := range klines1Min {
		if k1m.PriceInUsd > highest {
			highest = k1m.PriceInUsd
		}

		if !isRetrace {
			// 判断价格是否超过成本的50%, 如果是, 则定义为回调阶段
			if highest >= curPrice*1.5 {
				isRetrace = true
			}
		}

		// 亏50%或回调30% 定义为卖出
		if (k1m.PriceInUsd <= curPrice*0.5) || (isRetrace && k1m.PriceInUsd <= highest*0.7) {
			p.handleSellInfo(_pair, &k1m)
			return
		}
	}

	// 说明未卖掉, 表示持有
	p.handleSellInfo(_pair, nil)
}

// 卖出时间点, 如果kline为nil, 表示一直持有未卖出
func (p *HBPair) handleSellInfo(_pair *schema.Pair, kline *schema.KLine) {
	buyPrice := _pair.ReservedFloat // 取出回溯时的价格
	if buyPrice <= 0 {
		utils.Warnf("[  handleSellInfo] error buyPrice is: %v", buyPrice)
		return
	}

	var sellPrice float64
	var sellTime time.Time

	if kline != nil {
		sellPrice = kline.PriceInUsd
		sellTime = time.Unix(kline.UnixTime, 0)
	} else {
		sellPrice = _pair.PriceInUsd
		sellTime = time.Now()
	}

	earn := (sellPrice - buyPrice) / buyPrice * 100

	if kline != nil {
		fmt.Printf("%v, %.1f%%, %v, %v\n", _pair.PairName, earn, _pair.Address, sellTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("%v, %.1f%%, %v, %v\n", _pair.PairName, earn, _pair.Address, "hold")
	}
}

// 同时允许api调用, 略搓...
func (p *HBPair) HotPairSearch() ([]string, []string, []string) {
	// 先取出符合基础条件的pair, 并按24h交易量增加量排序
	pairs := p.GetBasicPairs()

	if p.traceTime.IsZero() {
		p.traceTime = time.Now()
	}

	for _, _pair := range pairs {
		// if p.checkPairHot1Hour(_pair) {
		// 	if !checkExistPair(_pair, p.hpairs1h) {
		// 		p.hpairs1h = append(p.hpairs1h, _pair)
		// 	}
		// }

		// if p.checkPairHot4Hour(_pair) {
		// 	if !checkExistPair(_pair, p.hpairs4h) {
		// 		p.hpairs4h = append(p.hpairs4h, _pair)
		// 	}
		// }

		if p.checkPairHot1Day(_pair) {
			if !checkExistPair(_pair, p.hpairs1d) {
				p.hpairs1d = append(p.hpairs1d, _pair)
			}
		}
	}

	var h1 []string
	var h4 []string
	var h24 []string

	for _, h := range p.hpairs1h {
		h1 = append(h1, h.Address)
	}
	for _, h := range p.hpairs4h {
		h4 = append(h4, h.Address)
	}
	for _, h := range p.hpairs1d {
		h24 = append(h24, h.Address)
	}

	return h1, h4, h24
}

func (p *HBPair) printPairs(hpairs []*schema.Pair) {
	utils.Infof("\n[ printPairs ] hpairs len: %v", len(hpairs))

	for _, pair := range hpairs {
		fmt.Printf("%v, %v\n", pair.Address, pair.PairName)
	}
}

// 其他基础条件已满足, 需判断k线相关逻辑
func (p *HBPair) checkPairHot1Hour(_pair *schema.Pair) bool {
	now := p.traceTime
	days := 1

	// 柱子是升序排列
	klines1Hour := kline.Get1HourKlineWithFullGenerated(_pair.Address, now, days, p.Set.DB.Database(config.DatabaseName))

	if len(klines1Hour) < 1*24 {
		utils.Debugf("[ checkPairHot1Hour ] klines1Hour len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}

	// 去掉最新的一根柱子, 因为他可能还未跑完, 非整点
	// 同时只取最新15根柱子
	klines1Hour = klines1Hour[len(klines1Hour)-16 : len(klines1Hour)-2]

	var volumeCur, priceCur float64
	klines1Min := kline.Get1MinKlineWithFullGenerated(_pair.Address, now, 1, p.Set.DB.Database(config.DatabaseName))
	if len(klines1Min) < 60 {
		utils.Warnf("Get1MinKlineWithFullGenerated len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}
	for _, kl1m := range klines1Min {
		volumeCur += kl1m.VolumeInUsd
	}
	priceCur = klines1Min[len(klines1Min)-1].PriceInUsd
	_pair.ReservedFloat = priceCur

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
	if volumeIncreament < p.Set.Config.VolumeIncrement {
		utils.Warnf("[ checkPairHot1Hour ] VolumeIncrement less, return. pairName: %v, addr: %v, volumeIncreament: %v", _pair.PairName, _pair.Address, volumeIncreament)
		return false
	}

	if priceIncreament < p.Set.Config.MinPriceIncreament {
		utils.Warnf("[ checkPairHot1Hour ] PriceIncreament less, return. pairName: %v, addr: %v, priceIncreament: %v, priceCur: %v, priceMid: %v", _pair.PairName, _pair.Address, priceIncreament, priceCur, priceMid)
		return false
	}

	// 1. 需要最新柱子的volume是历史柱子中最大的
	// 2. 同时需要最新价格与历史价格相比, 最大差值不超过 x%
	// 3. 历史柱子中不能出现 0 的情况
	for _, kline := range klines1Hour {
		if volumeCur < kline.VolumeInUsd {
			utils.Warnf("[ checkPairHot1Hour ] volume not the biggest, return. pairName: %v, addr: %v, volumeCur: %v, kline.VolumeInUsd: %v", _pair.PairName, _pair.Address, volumeCur, kline.VolumeInUsd)
			return false
		}

		if math.Abs(priceCur-kline.PriceInUsd)/priceCur > p.Set.Config.MaxPriceIncreament {
			utils.Warnf("[ checkPairHot1Hour ] bigger than MaxPriceIncreament, return. pairName: %v, addr: %v, priceCur: %v, kline.PriceInUsd: %v", _pair.PairName, _pair.Address, priceCur, kline.PriceInUsd)
			return false
		}

		if kline.VolumeInUsd <= 0 {
			utils.Warnf("[ checkPairHot1Hour ] volume is 0, return. pairName: %v, addr: %v, volumeCur: %v, kline.VolumeInUsd: %v", _pair.PairName, _pair.Address, volumeCur, kline.VolumeInUsd)
			return false
		}
	}

	utils.Infof("[ checkPairHot1Hour ] pass. pairName: %v, addr: %v, volumeInc: %v, priceInc: %v", _pair.PairName, _pair.Address, volumeIncreament, priceIncreament)

	return true
}

func (p *HBPair) checkPairHot4Hour(_pair *schema.Pair) bool {
	now := p.traceTime
	days := 3

	// 柱子是升序排列
	klines1Hour := kline.Get1HourKlineWithFullGenerated(_pair.Address, now, days, p.Set.DB.Database(config.DatabaseName))

	if len(klines1Hour) < 3*24 {
		utils.Warnf("[ checkPairHot4Hour ] klines1Hour len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}

	// 去掉最新的一根柱子, 因为他可能还未跑完, 非整点
	klines1Hour = klines1Hour[len(klines1Hour)-61 : len(klines1Hour)-2]

	// 需要通过分钟线取最近4小时的交易量
	var volumeCur, priceCur float64
	volumeCur = klines1Hour[len(klines1Hour)-1].VolumeInUsd + klines1Hour[len(klines1Hour)-2].VolumeInUsd + klines1Hour[len(klines1Hour)-3].VolumeInUsd + klines1Hour[len(klines1Hour)-4].VolumeInUsd
	priceCur = klines1Hour[len(klines1Hour)-1].PriceInUsd
	_pair.ReservedFloat = priceCur

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
	if volumeIncreament < p.Set.Config.VolumeIncrement {
		utils.Warnf("[ checkPairHot4Hour ] VolumeIncrement less, return. pairName: %v, addr: %v, volumeIncreament: %v", _pair.PairName, _pair.Address, volumeIncreament)
		return false
	}

	if priceIncreament < p.Set.Config.MinPriceIncreament || priceIncreament > p.Set.Config.MaxPriceIncreament {
		utils.Warnf("[ checkPairHot4Hour ] PriceIncreament less, return. pairName: %v, addr: %v, priceIncreament: %v, priceCur: %v, priceMid: %v", _pair.PairName, _pair.Address, priceIncreament, priceCur, priceMid)
		return false
	}

	// 需要最新柱子的volume是历史柱子中最大的
	var candiskVolume float64
	for ind, kline := range klines1Hour {
		candiskVolume += kline.VolumeInUsd

		if ind%4 == 0 && ind > 0 {
			if math.Abs(priceCur-kline.PriceInUsd)/priceCur > p.Set.Config.MaxPriceIncreament {
				utils.Warnf("[ checkPairHot4Hour ] bigger than MaxPriceIncreament, return. pairName: %v, addr: %v, priceCur: %v, kline.PriceInUsd: %v", _pair.PairName, _pair.Address, priceCur, kline.PriceInUsd)
				return false
			}

			if candiskVolume <= 0 {
				utils.Warnf("[ checkPairHot4Hour ] volume is 0, return. pairName: %v, addr: %v, volumeCur: %v, kline.VolumeInUsd: %v", _pair.PairName, _pair.Address, volumeCur, kline.VolumeInUsd)
				return false
			}

			candiskVolume = 0
		}

		if volumeCur < candiskVolume {
			utils.Warnf("[ checkPairHot4Hour ] volume not the biggest, return. pairName: %v, addr: %v, volumeCur: %v, candiskVolume: %v", _pair.PairName, _pair.Address, volumeCur, candiskVolume)
			return false
		}
	}

	utils.Infof("[ checkPairHot4Hour ] pass. pairName: %v, addr: %v, volumeInc: %v, priceInc: %v", _pair.PairName, _pair.Address, volumeIncreament, priceIncreament)

	return true
}

func (p *HBPair) checkPairHot1Day(_pair *schema.Pair) bool {
	now := p.traceTime
	days := 16

	// 柱子是升序排列
	klines1Hour := kline.Get1HourKlineWithFullGenerated(_pair.Address, now, days, p.Set.DB.Database(config.DatabaseName))

	if len(klines1Hour) < 16*24 {
		utils.Warnf("[ checkPairHot1Day ] klines1Hour len too less, return. pairName: %v, addr: %v", _pair.PairName, _pair.Address)
		return false
	}

	// 去掉最新的一根柱子, 因为他可能还未跑完, 非整点
	klines1Hour = klines1Hour[len(klines1Hour)-361 : len(klines1Hour)-2]

	var volumeCur, priceCur float64
	j := 0
	for i := len(klines1Hour) - 1; i > 0 && j < 24; i-- {
		volumeCur += klines1Hour[i].VolumeInUsd
		j++
	}
	priceCur = klines1Hour[len(klines1Hour)-1].PriceInUsd
	_pair.ReservedFloat = priceCur

	if volumeCur == 0 || priceCur == 0 {
		utils.Warnf("[ checkPairHot1Day ] current data is zero. return. pairName: %v, addr: %v, volumeCur: %v, priceCur: %v", _pair.PairName, _pair.Address, volumeCur, priceCur)
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
	if volumeIncreament < p.Set.Config.VolumeIncrement {
		utils.Warnf("[ checkPairHot1Day ] VolumeIncrement less, return. pairName: %v, addr: %v, volumeIncreament: %v", _pair.PairName, _pair.Address, volumeIncreament)
		return false
	}

	if priceIncreament < p.Set.Config.MinPriceIncreament || priceIncreament > p.Set.Config.MaxPriceIncreament {
		utils.Warnf("[ checkPairHot1Day ] PriceIncreament less, return. pairName: %v, addr: %v, priceIncreament: %v, priceCur: %v, priceMid: %v", _pair.PairName, _pair.Address, priceIncreament, priceCur, priceMid)
		return false
	}

	var candiskVolume float64
	for ind, kline := range klines1Hour {
		candiskVolume += kline.VolumeInUsd

		if ind%24 == 0 && ind > 0 {
			// 当前价格与历史任何价格之差不能大于一定比例
			if math.Abs(priceCur-kline.PriceInUsd)/priceCur > p.Set.Config.MaxPriceIncreament {
				utils.Warnf("[ checkPairHot1Day ] bigger than MaxPriceIncreament, return. pairName: %v, addr: %v, priceCur: %v, kline.PriceInUsd: %v", _pair.PairName, _pair.Address, priceCur, kline.PriceInUsd)
				return false
			}

			// 历史柱子中不允许出现为0交易量
			if candiskVolume <= 0 {
				utils.Warnf("[ checkPairHot1Day ] volume is 0, return. pairName: %v, addr: %v, volumeCur: %v, kline.VolumeInUsd: %v", _pair.PairName, _pair.Address, volumeCur, kline.VolumeInUsd)
				return false
			}

			candiskVolume = 0
		}

		// 需要最新柱子的volume是历史柱子中最大的
		if volumeCur < candiskVolume {
			utils.Warnf("[ checkPairHot1Day ] volume not the biggest, return. pairName: %v, addr: %v, volumeCur: %v, candiskVolume: %v", _pair.PairName, _pair.Address, volumeCur, candiskVolume)
			return false
		}

	}

	utils.Infof("[ checkPairHot1Day ] pass. pairName: %v, addr: %v, volumeInc: %v, priceInc: %v", _pair.PairName, _pair.Address, volumeIncreament, priceIncreament)

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
		"$gte": p.Set.Config.MinPairLiquidity,
	}

	// 价格最小涨幅要求, 减少数据量
	filter["priceChangeIn24h"] = bson.M{
		"$gte": 0,
	}
	// filter["$and"] = []bson.M{
	// 	{"priceChangeIn24h": bson.M{"$gte": 0}},
	// 	{"priceChangeIn24h": bson.M{"$lte": 0.5}},
	// }

	// pair创建时长要求
	date := time.Now().Add(-time.Duration(p.Set.Config.MinPairCreatAge) * time.Second)
	filter["firstAddPoolTime"] = bson.M{
		"$lte": date,
	}

	// 非通缩币
	filter["mainTokenHackType"] = bson.M{
		"$lte": 2,
	}

	// 最近1小时内必须更新过, 也就是必须有交易
	date = time.Now().Add(-1 * time.Hour)
	filter["updatedAt"] = bson.M{
		"$gte": date,
	}

	// 基础排序策略
	limit := int64(1000)
	options := &options.FindOptions{Limit: &limit}

	sort := bson.D{bson.E{Key: "txNumIn24h", Value: -1}}
	options = options.SetSort(sort)

	db := p.Set.DB.Database(config.DatabaseName)
	pairs, _, err := pair.GetHotPairs(options, &filter, db)
	if err != nil {
		utils.Errorf("[ GetBasicPairs ] GetHotPairs err: %v", err)
		return nil
	}

	return pairs
}

func checkExistPair(target *schema.Pair, pairs []*schema.Pair) bool {
	targetMainToken := utils.GetMainToken(target.Token0, target.Token1)

	for _, element := range pairs {
		pairMainToken := utils.GetMainToken(element.Token0, element.Token1)

		if targetMainToken == pairMainToken {
			utils.Debugf("[ checkExistPair ] same. token: %v, target: %v, element: %v", targetMainToken, target.PairName, element.PairName)
			return true
		}

	}

	return false
}
