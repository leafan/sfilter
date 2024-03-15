package handler

import (
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
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
@date: 20240304

买入条件(取与):
1. 过滤掉通缩币
2. pair创建时间小于1小时且池子大于1万美金
3. 最高价/最新价 小于2
4. 进入小时tx前5
5. 每小时的整点跑一次

end: 统计符合前4个条件的次数并打印

@date: 20240305
卖点:
1. 涨了5倍卖出
2. 每10min检查tx跌破前5

*/

type HNPair struct {
	Set *Setting
}

func (p *HNPair) Run() {
	// { // test
	// 	p.PolicyLogic()
	// 	return
	// }

	c := cron.New()

	spec_min := "20 0/10 * * * *"
	c.AddFunc(spec_min, func() {
		p.GetHotNewPairSellSignal()
	})

	spec_hour := "30 0 * * * *"
	c.AddFunc(spec_hour, func() {
		p.GetHotNewPairBuySignal()
	})

	c.Start()

	select {}
}

// for test
func (p *HNPair) PolicyLogic() {
	p.GetHotNewPairSellSignal() // 先检查卖的, 避免本次新加入的也被检查
	p.GetHotNewPairBuySignal()
}

func (p *HNPair) GetHotNewPairBuySignal() {
	pairs := p.FindHotNewPairs(5)

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
		trade := schema.BiTrade{
			PairAddress:   _pair.Address,
			PairName:      _pair.PairName,
			PairLiquidity: _pair.LiquidityInUsd,
			PairAge:       time.Since(_pair.FirstAddPoolTime),

			BuyPrice:  _pair.PriceInUsd,
			BuyTime:   time.Now(),
			SortRank:  _pair.ReservedInt,
			TxNumIn1h: _pair.TxNumIn1h,
		}

		p.buyTrade(&trade)

	}
}

func (p *HNPair) checkPairValidity(_pair *schema.Pair) bool {
	// 先检测是否为坑人币
	if _pair.Type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
		hackType := chain.GetUniV2PoolTokenHackType(_pair)
		if hackType > 3 {
			// 说明是坑人币
			utils.Errorf("[ checkPairValidity ] hack pair: %v. Pair: %v, address: %v", hackType, _pair.PairName, _pair.Address)
			return false
		}
	}

	// pair创建时间小于24小时
	if time.Since(_pair.FirstAddPoolTime) > time.Duration(1*time.Hour) {
		utils.Warnf("[ checkPairValidity ] pair too old. pair: %v, addr: %v, duration: %v", _pair.PairName, _pair.Address, time.Since(_pair.FirstAddPoolTime))
		return false
	}

	var highest float64

	now := time.Now()
	klines1Min := kline.Get1MinKlineWithFullGenerated(_pair.Address, now, 24, p.Set.DB.Database(config.DatabaseName))
	for _, k1m := range klines1Min {
		if k1m.HighPrice > highest {
			highest = k1m.HighPrice
		}
	}

	// 最高价/最新价 小于x
	if highest/_pair.Price >= 2 {
		utils.Warnf("[ checkPairValidity ] pair price ratio too high. pair: %v, addr: %v, highest: %v, current: %v", _pair.PairName, _pair.Address, highest, _pair.Price)
		return false
	}

	utils.Infof("[ checkPairValidity ] pass.. pair: %v", _pair.PairName)

	return true
}

func (p *HNPair) GetHotNewPairSellSignal() {
	trades := p.getUnSoldTrades()
	topN := p.FindHotNewPairs(5)

	for _, trade := range trades {
		isSell := p.checkSellSignal(&trade, topN)
		if isSell {
			p.sellTrade(&trade)
		}
	}
}

func (p *HNPair) checkSellSignal(trade *schema.BiTrade, topN []*schema.Pair) bool {
	traceHours := 2
	// traceHours := int(time.Since(trade.BuyTime).Hours()) + 2
	klines1Min := kline.Get1MinKlineWithFullGenerated(trade.PairAddress, time.Now(), traceHours, p.Set.DB.Database(config.DatabaseName))

	var k1m schema.KLine
	for _, k1m = range klines1Min {
		// 先要确认当前时间大于买入时间
		if k1m.UnixTime <= trade.BuyTime.Unix() {
			continue
		}

		// 1. 涨了x倍卖出
		earnX := float64(5)
		if k1m.PriceInUsd >= trade.BuyPrice*earnX {
			utils.Infof("PairName: %v, earn %vX now.", trade.PairName, earnX)

			trade.SellTime = time.Unix(k1m.UnixTime, 0)
			trade.SellPrice = k1m.PriceInUsd
			trade.SellReason = fmt.Sprintf("赚%.0f倍卖", earnX)
			return true
		}

		// 跌破成本50%卖出
		// if k1m.PriceInUsd <= trade.BuyPrice*0.5 {
		// 	utils.Warnf("PairName: %v, lose 50%% now", trade.PairName)

		// 	trade.SellTime = time.Unix(k1m.UnixTime, 0)
		// 	trade.SellPrice = k1m.PriceInUsd
		// 	trade.SellReason = fmt.Sprintf("亏50%%卖")
		// 	return true
		// }
	}

	// 检查热度是否跌出了前x
	for _, _pair := range topN {
		if trade.PairAddress == _pair.Address {
			return false // 还在前x, 继续持有
		}
	}

	utils.Infof("PairName: %v, out of topx.", trade.PairName)
	trade.SellTime = time.Now()
	if len(klines1Min) > 0 {
		trade.SellPrice = klines1Min[len(klines1Min)-1].PriceInUsd
	} else {
		trade.SellPrice = 0
	}
	trade.SellReason = "TX 跌出前10"

	return true
}

func (p *HNPair) sellTrade(trade *schema.BiTrade) {
	trade.EarnRatio = (trade.SellPrice - trade.BuyPrice) / trade.BuyPrice
	trade.HoldTime = trade.SellTime.Sub(trade.BuyTime)
	trade.Status = 1

	// 保存到db
	wiser.UpdateBiTrade(trade, p.Set.DB)

	msg := fmt.Sprintf("<font color=\"warning\">[ **Sell** ]</font>\nPair: [%v](https://www.dextools.io/app/cn/ether/pair-explorer/%v)\nEarn: <font color=\"comment\">%.2f%%</font>\nReason: <font color=\"comment\">**%v**</font>\nSellPrice: %v\nBuyTime: %v\nBuyRank: %v\nHoldTime: %v\nLiquidity: %v", trade.PairName, trade.PairAddress, trade.EarnRatio*100, trade.SellReason, trade.SellPrice, trade.BuyTime.In(time.FixedZone("UTC+8", 8*60*60)).Format("01-02 15:04"), trade.SortRank, utils.ReadibleDuration(trade.HoldTime), utils.HumanizeNumber(trade.PairLiquidity))

	utils.SendWecommBot(p.Set.Config.HotPairHookUrl, msg)
}

func (p *HNPair) getUnSoldTrades() []schema.BiTrade {
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

func (p *HNPair) buyTrade(trade *schema.BiTrade) {
	utils.Infof("[ buyTrade ] buy now. PairName: %v", trade.PairName)

	// 保存到db
	wiser.SaveBiTrade(trade, p.Set.DB)

	// 获取该pair符合要求的次数
	// buyCounts := p.GetPairBuyCounts(trade.PairAddress)

	// 发送消息到wx
	msg := fmt.Sprintf("<font color=\"info\">[ **Buy** ]</font>\nPair: [%v](https://www.dextools.io/app/cn/ether/pair-explorer/%v)\nLiquidity: $%v\nTxNumIn1h: %v\nAge: %v\nPrice: %v\nRank: %v\n", trade.PairName, trade.PairAddress, utils.HumanizeNumber(trade.PairLiquidity), trade.TxNumIn1h, utils.ReadibleDuration(trade.PairAge), trade.BuyPrice, trade.SortRank)

	utils.SendWecommBot(p.Set.Config.HotPairHookUrl, msg)
}

func (p *HNPair) GetPairBuyCounts(pair string) int64 {
	filter := bson.M{}
	filter["pairAddress"] = pair

	options := &options.FindOptions{}
	_, count, _ := wiser.GetBiTrades(options, &filter, p.Set.DB.Database(config.DatabaseName))

	return count
}

func (p *HNPair) FindHotNewPairs(topN int64) []*schema.Pair {
	db := p.Set.DB.Database(config.DatabaseName)
	filter := bson.M{}

	// 非通缩币
	filter["mainTokenHackType"] = bson.M{
		"$lte": 2,
	}

	// 最近1小时内必须更新过, 也就是必须有交易
	date := time.Now().Add(-1 * time.Hour)
	filter["tradeInfoUpdatedAt"] = bson.M{
		"$gte": date,
	}

	// 池子 tvl 要求
	filter["liquidityInUsd"] = bson.M{
		"$gte": 10000,
	}

	// 把常见的几个 大pair 排除掉
	filter["$and"] = []bson.M{
		{"address": bson.M{"$ne": "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"}}, // WETH/USDC_UniV3
		{"address": bson.M{"$ne": "0x11b815efB8f581194ae79006d24E0d814B7697F6"}}, // WETH/USDT_UniV3
		{"address": bson.M{"$ne": "0x0d4a11d5EEaaC28EC3F61d100daF4d40471f1852"}}, // WETH/USDT_UniV2
		{"address": bson.M{"$ne": "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc"}}, // WETH/USDC_UniV2
		{"address": bson.M{"$ne": "0xc7bBeC68d12a0d1830360F8Ec58fA599bA1b0e9b"}}, // WETH/USDT_UniV3
	}

	// 取前x名
	options := &options.FindOptions{Limit: &topN}

	sort := bson.D{}

	// 根据 txNumIn1h 倒排
	sort = append(sort, bson.E{Key: "txNumIn1h", Value: -1}, bson.E{Key: "UpdateAt", Value: -1})
	options = options.SetSort(sort)

	info, _, err := pair.GetHotPairs(options, &filter, db)

	utils.Debugf("[ FindHotNewPairs ] top x print. err: %v", err)
	for ind, _pair := range info {
		fmt.Printf("%v: pair: %v, address: %v\n", ind+1, _pair.PairName, _pair.Address)
	}

	return info
}
