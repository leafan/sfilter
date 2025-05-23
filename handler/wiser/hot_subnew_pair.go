package handler

import (
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/kline"
	"sfilter/services/pair"
	"sfilter/services/wiser"
	"sfilter/utils"
	"time"

	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
买入条件（取与）：
1. 过滤掉通缩币, 根据1小时tx数倒序排名, 取出前20名pair
2. 每5分钟取tx, 连续15分钟tx上涨, 且每5分钟都有至少1笔tx
3. 每5分钟取价格, 连续15分钟价格上涨, 且最新价格比15分钟前价格大于5%且小于50%

卖出条件（取或）：
1. 跌破成本50%或者高点回调幅度超过30%
2. 缩量涨(交易量连续3根5min柱子跌但价格比15min前高)
3. 放量跌(交易量连续3根5min柱子涨但价格比15min前低)


----- 废弃卖出条件:
1. 每5分钟取tx, 连续15分钟tx下降
2. 每5分钟取交易量, 连续15分钟下降
*/

type HSPair struct {
	Set *Setting
}

func (p *HSPair) Run() {
	// { // test
	// 	p.PolicyLogic()
	// 	return
	// }

	c := cron.New()
	spec := "20 0/5 * * * *" // 第20s开始执行, 每5min执行一次
	c.AddFunc(spec, func() {
		p.PolicyLogic()
	})
	c.Start()

	select {}
}

func (p *HSPair) PolicyLogic() {
	p.HotSubnewSellSignal() // 先检查卖的, 避免本次新加入的也被检查

	p.HotSubnewBuySignal()
}

func (p *HSPair) HotSubnewBuySignal() {
	pairs := p.FindHotSubnewPairs()

	var buyPairs []*schema.Pair

	for ind, _pair := range pairs {
		if p.checkPairValidity(_pair) {
			if !checkExistPair(_pair, buyPairs) {
				_pair.ReservedInt = ind + 1 // 把其排序参数写进去
				buyPairs = append(buyPairs, _pair)
			}
		}
	}

	// 针对选择出来的程序, 执行买入动作
	for _, _pair := range buyPairs {
		// 检查是否已经买入
		mainToken := utils.GetMainToken(_pair.Token0, _pair.Token1)
		_, err := wiser.GetToBeSoldBiTrade(mainToken, p.Set.DB)
		if err != nil {
			// 没找到, 或者说没买, 执行买入动作
			trade := schema.BiTrade{
				MainToken:     mainToken,
				PairAddress:   _pair.Address,
				PairName:      _pair.PairName,
				PairLiquidity: _pair.LiquidityInUsd,
				PairAge:       time.Since(_pair.FirstAddPoolTime),

				BuyPrice:  _pair.PriceInUsd,
				BuyTime:   time.Now(),
				BuyReason: _pair.ReservedString,
				SortRank:  _pair.ReservedInt,

				HighestPrice: _pair.PriceInUsd,
			}

			p.buyTrade(&trade)
		} else {
			utils.Warnf("[ HotSubnewBuySignal ] mainToken has been bought. pairName: %v, mainToken: %v", _pair.PairName, mainToken)
		}
	}
}

func (p *HSPair) checkPairValidity(_pair *schema.Pair) bool {
	// return true // test

	// 取2根小时柱子
	klines1Min := kline.Get1MinKlineWithFullGenerated(_pair.Address, time.Now(), 2, p.Set.DB.Database(config.DatabaseName))

	// 倒推15min, 每5min组一根柱子
	if len(klines1Min) < 2*60 {
		// utils.Warnf("[ checkPairValidity ] candisk is too less, len: %v, pairName: %v, pairAddr: %v", len(klines1Min), _pair.PairName, _pair.Address)
		return false
	}
	// 过滤掉最后一根柱子, 因为可能当前分钟还没结束, 柱子非完整
	klines1Min = klines1Min[:len(klines1Min)-1]

	length := len(klines1Min)
	tx3 := klines1Min[length-1].TxNum + klines1Min[length-2].TxNum + klines1Min[length-3].TxNum + klines1Min[length-4].TxNum + klines1Min[length-5].TxNum
	tx2 := klines1Min[length-6].TxNum + klines1Min[length-7].TxNum + klines1Min[length-8].TxNum + klines1Min[length-9].TxNum + klines1Min[length-10].TxNum
	tx1 := klines1Min[length-11].TxNum + klines1Min[length-12].TxNum + klines1Min[length-13].TxNum + klines1Min[length-14].TxNum + klines1Min[length-15].TxNum

	// 每5分钟取tx，连续15分钟tx上涨
	if !(tx1 > 0 && tx1 < tx2 && tx2 < tx3) {
		// utils.Debugf("[ checkPairValidity ] tx not pass, tx1: %v, tx2: %v, tx3: %v, pair: %v", tx1, tx2, tx3, _pair.PairName)
		return false
	}

	// 每5分钟取价格，连续15分钟价格上涨
	if !(klines1Min[length-1].PriceInUsd > klines1Min[length-6].PriceInUsd &&
		klines1Min[length-6].PriceInUsd > klines1Min[length-11].PriceInUsd &&
		klines1Min[length-11].PriceInUsd > klines1Min[length-16].PriceInUsd) {
		// utils.Debugf("[ checkPairValidity ] price increment not pass, pair: %v", _pair.PairName)
		return false
	}

	// 价格15min内涨幅
	if klines1Min[length-1].PriceInUsd < klines1Min[length-16].PriceInUsd*1.05 || klines1Min[length-1].PriceInUsd > klines1Min[length-16].PriceInUsd*1.5 {
		// utils.Debugf("[ checkPairValidity ] price not pass, pair: %v", _pair.PairName)
		return false
	}

	utils.Infof("[ checkPairValidity ] pass, pair: %v, tx1: %v, tx2: %v, tx3: %v", _pair.PairName, tx1, tx2, tx3)
	_pair.ReservedString = fmt.Sprintf("tx: %v->%v->%v", tx1, tx2, tx3)
	return true
}

func (p *HSPair) HotSubnewSellSignal() {
	// 取出所有未卖出的币, 逐个检查是否达到了卖出条件
	trades := p.getUnSoldTrades()

	for _, trade := range trades {
		isSell := p.checkSellSignal(&trade)
		if isSell {
			p.sellTrade(&trade)
		}
	}
}

func (p *HSPair) checkSellSignal(trade *schema.BiTrade) bool {
	// 取2根小时柱子
	klines1Min := kline.Get1MinKlineWithFullGenerated(trade.PairAddress, time.Now(), 2, p.Set.DB.Database(config.DatabaseName))

	// 倒推15min, 每5min组一根柱子
	if len(klines1Min) < 2*60 {
		// utils.Warnf("[ checkSellSignal ] candisk is too less, len: %v, pairName: %v, pairAddr: %v", len(klines1Min), trade.PairName, trade.PairAddress)
		return false
	}
	// 过滤掉最后一根柱子, 因为可能当前分钟还没结束, 柱子非完整
	klines1Min = klines1Min[:len(klines1Min)-1]

	length := len(klines1Min)
	// tx3 := klines1Min[length-1].TxNum + klines1Min[length-2].TxNum + klines1Min[length-3].TxNum + klines1Min[length-4].TxNum + klines1Min[length-5].TxNum
	// tx2 := klines1Min[length-6].TxNum + klines1Min[length-7].TxNum + klines1Min[length-8].TxNum + klines1Min[length-9].TxNum + klines1Min[length-10].TxNum
	// tx1 := klines1Min[length-11].TxNum + klines1Min[length-12].TxNum + klines1Min[length-13].TxNum + klines1Min[length-14].TxNum + klines1Min[length-15].TxNum

	// // 每5分钟取tx, 连续15分钟tx下降
	// if tx1 > tx2 && tx2 > tx3 {
	// 	utils.Infof("[ checkSellSignal ] PairName: %v, tx pass. tx1: %v, tx2: %v, tx3: %v", trade.PairName, tx1, tx2, tx3)
	// 	trade.SellReason = fmt.Sprintf("Tx decrease: %v->%v->%v", tx1, tx2, tx3)
	// 	trade.SellPrice = klines1Min[len(klines1Min)-1].PriceInUsd
	// 	return true
	// }

	vl3 := klines1Min[length-1].VolumeInUsd + klines1Min[length-2].VolumeInUsd + klines1Min[length-3].VolumeInUsd + klines1Min[length-4].VolumeInUsd + klines1Min[length-5].VolumeInUsd
	vl2 := klines1Min[length-6].VolumeInUsd + klines1Min[length-7].VolumeInUsd + klines1Min[length-8].VolumeInUsd + klines1Min[length-9].VolumeInUsd + klines1Min[length-10].VolumeInUsd
	vl1 := klines1Min[length-11].VolumeInUsd + klines1Min[length-12].VolumeInUsd + klines1Min[length-13].VolumeInUsd + klines1Min[length-14].VolumeInUsd + klines1Min[length-15].VolumeInUsd

	// 1. 缩量涨(交易量连续3根5min柱子跌但价格比15min前高)
	if vl1 > vl2 && vl2 > vl3 && klines1Min[len(klines1Min)-16].PriceInUsd < klines1Min[len(klines1Min)-1].PriceInUsd {
		utils.Infof("[ checkSellSignal ] PairName: %v, volume pass. vl1: %v, vl2: %v, vl3: %v", trade.PairName, vl1, vl2, vl3)
		trade.SellReason = fmt.Sprintf("缩量涨: %.1f->%.1f->%.1f", vl1, vl2, vl3)
		trade.SellPrice = klines1Min[len(klines1Min)-1].PriceInUsd
		return true
	}

	// 2. 放量跌(交易量连续3根5min柱子涨但价格比15min前低)
	if vl1 < vl2 && vl2 < vl3 && klines1Min[len(klines1Min)-1].PriceInUsd < klines1Min[len(klines1Min)-16].PriceInUsd {
		utils.Infof("[ checkSellSignal ] PairName: %v, volume pass. vl1: %v, vl2: %v, vl3: %v", trade.PairName, vl1, vl2, vl3)
		trade.SellReason = fmt.Sprintf("放量跌: %.1f->%.1f->%.1f", vl1, vl2, vl3)
		trade.SellPrice = klines1Min[len(klines1Min)-1].PriceInUsd
		return true
	}

	// 3: 跌破成本50%或者高点回调幅度超过30%
	var isRetrace bool // 是否属于回调阶段
	j := 0             // 最近5-6min是否符合卖出条件
	for i := len(klines1Min) - 1; i > 0 && j < 6; i-- {
		k1m := klines1Min[i]
		j++

		if k1m.PriceInUsd > trade.HighestPrice {
			trade.HighestPrice = k1m.PriceInUsd

			// 更新到db, 下次使用
			wiser.UpdateBiTrade(trade, p.Set.DB)
		}

		if !isRetrace {
			// 判断最高价格是否超过成本的50%, 如果是, 则定义为回调阶段
			if trade.HighestPrice >= trade.BuyPrice*1.5 {
				isRetrace = true
			}
		}

		// 亏50%卖出
		if k1m.PriceInUsd <= trade.BuyPrice*0.5 {
			utils.Infof("[ checkSellSignal ] PairName: %v, lose money. Current Price: %v, trade.BuyPrice: %v", trade.PairName, k1m.PriceInUsd, trade.BuyPrice)

			trade.SellReason = fmt.Sprintf("跌破50%%. Buy: %v, Current: %v", trade.BuyPrice, k1m.PriceInUsd)
			trade.SellPrice = k1m.PriceInUsd
			return true
		}

		// 回调30%为卖出
		if isRetrace && k1m.PriceInUsd <= trade.HighestPrice*0.7 {
			utils.Infof("[ checkSellSignal ] PairName: %v, retrace pass. Current: %v,HighestPrice", trade.PairName, k1m.PriceInUsd, trade.HighestPrice)

			trade.SellReason = fmt.Sprintf("回调卖出. Current: %v, Highest: %v", k1m.PriceInUsd, trade.HighestPrice)
			trade.SellPrice = k1m.PriceInUsd
			return true
		}

	}

	return false
}

func (p *HSPair) buyTrade(trade *schema.BiTrade) {
	utils.Infof("[ buyTrade ] buy now. PairName: %v", trade.PairName)

	// 保存到db
	wiser.SaveBiTrade(trade, p.Set.DB)

	// 发送消息到wx
	msg := fmt.Sprintf("<font color=\"info\">[ **Buy** ]</font>\nPair: [%v](https://www.dextools.io/app/cn/ether/pair-explorer/%v)\nLiquidity: $%.1f K\nAge: %v\nPrice: %v\nRank: %v\nBuyReason: %v", trade.PairName, trade.PairAddress, trade.PairLiquidity/1000, utils.ReadibleDuration(trade.PairAge), trade.BuyPrice, trade.SortRank, trade.BuyReason)
	utils.SendWecommBot(p.Set.Config.HotPairHookUrl, msg)
}

func (p *HSPair) sellTrade(trade *schema.BiTrade) {
	// utils.Infof("[ sellTrade ] sell now. PairName: %v", trade.PairName)
	trade.SellTime = time.Now()

	trade.EarnRatio = (trade.SellPrice - trade.BuyPrice) / trade.BuyPrice
	trade.HoldTime = time.Since(trade.BuyTime)
	trade.Status = 1 // 表示该币可以重新买入

	// 保存到db
	wiser.UpdateBiTrade(trade, p.Set.DB)

	mainTokenCount := p.getSoldTradesByMainToken(trade.MainToken)

	// 发送消息到wx
	msg := fmt.Sprintf("<font color=\"warning\">[ **Sell** ]</font>\nPair: [%v](https://www.dextools.io/app/cn/ether/pair-explorer/%v)\nEarn: <font color=\"comment\">%.2f%%</font>\nReason: <font color=\"comment\">**%v**</font>\nSellPrice: %v\nBuyTime: %v\nBuyRank: %v\nHoldTime: %v\nBuyCounts: %v\nLiquidity: $%.1f K\nAge: %v", trade.PairName, trade.PairAddress, trade.EarnRatio*100, trade.SellReason, trade.SellPrice, trade.BuyTime.In(time.FixedZone("UTC+8", 8*60*60)).Format("01-02 15:04"), trade.SortRank, trade.HoldTime.String(), mainTokenCount, trade.PairLiquidity/1000, utils.ReadibleDuration(trade.PairAge))

	utils.SendWecommBot(p.Set.Config.HotPairHookUrl, msg)
}

func (p *HSPair) getSoldTradesByMainToken(mainToken string) int64 {
	filter := bson.M{}
	filter["status"] = 1
	filter["mainToken"] = mainToken

	options := &options.FindOptions{}
	_, count, _ := wiser.GetBiTrades(options, &filter, p.Set.DB.Database(config.DatabaseName))

	return count
}

func (p *HSPair) getUnSoldTrades() []schema.BiTrade {
	db := p.Set.DB.Database(config.DatabaseName)

	filter := bson.M{}
	filter["status"] = 0

	options := &options.FindOptions{}

	trades, _, err := wiser.GetBiTrades(options, &filter, db)
	if err != nil {
		utils.Errorf("[ getUnSoldTrades ] GetBiTrades failed: %v", err)
	}

	utils.Debugf("[ getUnSoldTrades ] GetBiTrades len: %v", len(trades))
	return trades
}

func (p *HSPair) FindHotSubnewPairs() []*schema.Pair {
	db := p.Set.DB.Database(config.DatabaseName)
	filter := bson.M{}

	// 非通缩币
	filter["mainTokenHackType"] = bson.M{
		"$lte": 2,
	}

	// 最近1小时内必须更新过, 也就是必须有交易
	date := time.Now().Add(-1 * time.Hour)
	filter["updatedAt"] = bson.M{
		"$gte": date,
	}

	limit := int64(20)
	options := &options.FindOptions{Limit: &limit}

	sort := bson.D{}

	// 根据 txNumIn1h 倒排
	sort = append(sort, bson.E{Key: "txNumIn1h", Value: -1}, bson.E{Key: "UpdateAt", Value: -1})
	options = options.SetSort(sort)

	info, count, err := pair.GetHotPairs(options, &filter, db)
	utils.Debugf("[ FindHotSubnewPairs ] len: %v, err: %v", count, err)

	return info
}
