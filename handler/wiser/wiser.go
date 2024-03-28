package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/pair"
	"sfilter/services/wiser"
	"sfilter/utils"
	"time"

	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Wiser struct {
	set *Setting

	dealInspect  bool
	wiserInspect bool

	epoch string
}

func (w *Wiser) Run() {
	{
		// w.Save1HourTopRank()
		// return
	}

	c := cron.New()

	spec_hour := "20 0/5 * * * *"
	c.AddFunc(spec_hour, func() {
		w.Save1HourTopRank()
	})

	c.Start()

	select {}
}

func (w *Wiser) fillRankWithPair(rank *schema.HotPairRank, _pair *schema.Pair) {
	rank.MainToken = utils.GetMainToken(_pair.Token0, _pair.Token1)
	rank.PairAddress = _pair.Address
	rank.PairName = _pair.PairName

	rank.PairLiquidity = _pair.LiquidityInUsd
	rank.PairAge = time.Since(_pair.FirstAddPoolTime)

	rank.TradeInfoForPair = _pair.TradeInfoForPair

	rank.CreatedAt = time.Now()
}

func (w *Wiser) Save1HourTopRank() {
	pairs := w.FindTopXPairs(50, "txNumIn1h")
	now := time.Now()

	var rankList []interface{}
	for ind, _pair := range pairs {
		pKey := fmt.Sprintf("%v_%v_%v_%v_%v", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())

		rank := schema.HotPairRank{
			PeriodKey:         pKey,
			PeriodKeyWithPair: fmt.Sprintf("%v_%v", pKey, _pair.Address),
			SortRank:          ind + 1,
		}

		// utils.Tracef("[ Save1HourTopRank ] pairname: %v, txnum: %v", _pair.PairName, _pair.TxNumIn1h)
		w.fillRankWithPair(&rank, _pair)

		rankList = append(rankList, rank)
	}

	// 批量存储
	err := wiser.SaveTopRanks(rankList, w.set.DB)
	if err != nil {
		utils.Errorf("[ Save1HourTopRank ] SaveTopRanks error: %v", err)
	}
}

func (w *Wiser) Test(accounts []string) {
	var contract, notContract int
	for _, account := range accounts {
		isContract := chain.IsContract(account)
		utils.Infof("[ Test ] account: %v is contract: %v", account, isContract)

		if isContract > 0 {
			contract++
		} else {
			notContract++
		}
	}

	utils.Fatalf("[ Test ] contract: %v, notContract: %v", contract, notContract)
}

func (w *Wiser) WiserSearcher() {
	var accounts []string
	if w.set.Config.DebugAccount != "" {
		accounts = append(accounts, w.set.Config.DebugAccount)
	} else {
		var err error
		accounts, err = wiser.GetActiveAccounts(w.set.Config.AccountActiveSeconds, w.set.DB)
		if err != nil {
			utils.Fatalf("[ WiserSearcher ] GetActiveAccounts error:  %v", err)
			return
		}
	}
	utils.Infof("[ WiserSearcher ] accounts len: %v, they are: %v", len(accounts), accounts)

	if w.dealInspect {
		if !w.set.Config.DebugMode && w.set.Config.DebugAccount == "" {
			wiser.ResetDealCollection(w.set.DB)
		}
		for _, account := range accounts {
			w.InspectBiDeals(account)
		}
	}
	if w.wiserInspect {
		w.epoch = time.Now().Format("20060102")
		if !w.set.Config.DebugMode && w.set.Config.DebugAccount == "" {
			wiser.ResetWiserEpochData(w.set.DB, w.epoch)
		}

		for _, account := range accounts {
			w.InspectAccount(account)
		}

		if !w.set.Config.DebugMode && w.set.Config.DebugAccount == "" {
			jsonString, _ := json.Marshal(w.set.Config)
			config := &schema.WiserDBConfig{
				Epoch:  w.epoch,
				Config: string(jsonString),
			}
			wiser.UpdateWiserConfig(config, w.set.DB) // 更新本次epoch
		}
	}

}

func (w *Wiser) FindTopXPairs(topN int64, sortKey string) []*schema.Pair {
	db := w.set.DB.Database(config.DatabaseName)
	filter := bson.M{}

	// 最近1小时内必须更新过, 也就是必须有交易
	date := time.Now().Add(-1 * time.Hour)
	filter["tradeInfoUpdatedAt"] = bson.M{
		"$gte": date,
	}

	// 池子 tvl 要求, 因为是取排名, 所以要求1000u至少
	filter["liquidityInUsd"] = bson.M{
		"$gte": 1000,
	}

	// 取前x名
	options := &options.FindOptions{Limit: &topN}

	sort := bson.D{}

	// 倒排
	sort = append(sort, bson.E{Key: sortKey, Value: -1}, bson.E{Key: "UpdateAt", Value: -1})
	options = options.SetSort(sort)

	info, _, err := pair.GetHotPairs(options, &filter, db)

	utils.Debugf("[ FindTopXPairs ] top x print. len: %v, err: %v", len(info), err)
	for ind, _pair := range info {
		fmt.Printf("%v: pair: %v, address: %v\n", ind+1, _pair.PairName, _pair.Address)
	}

	return info
}

// 分析某个账号是否是优秀地址
func (w *Wiser) InspectAccount(account string) {
	utils.Infof("[ InspectAccount ] account: %v", account)

	// 先取出该地址所有deals
	deals, err := wiser.GetAccountAllDeals(account, w.set.DB)
	if err != nil {
		utils.Errorf("[ InspectAccount ] failed to GetAccountAllDeals: %v", err)
		return
	}

	if w.set.Config.DebugAccount == "" && len(deals) < int(w.set.Config.DealThresholdPerMon) {
		utils.Debugf("[ InspectAccount ] GetAccountAllDeals failed or too less deals.  account: %v, len: %v", account, len(deals))
		return
	}

	// 统计wiser数据
	_wiser := w.inspectAccountByDeals(account, deals)

	// 获取余额
	ethBalance, err1 := chain.GetAccountEthBalance(account)
	wethBalance, err2 := chain.GetAccountWEthBalance(account)
	if err1 != nil || err2 != nil {
		utils.Errorf("[ InspectAccount ] GetAccountEthBalance failed. err1: %v, err2: %v", err1, err2)
		return
	}
	_wiser.EthBalance = ethBalance + wethBalance

	// 判断是否为合约
	_wiser.IsContract = chain.IsContract(_wiser.Address)

	isValid := w.isWiserNeedBePicked(&_wiser)
	if isValid {
		wiser.SaveWiser(&_wiser, w.set.DB)
	}

	// debug
	wiser.PrintWiser(&_wiser)
}

func (w *Wiser) isWiserNeedBePicked(wiser *schema.Wiser) bool {
	w.updateWiserWeight(wiser)

	// 如果余额过小, 或者其他条件不符合, 则不保存
	if wiser.EthBalance < w.set.Config.WiserMinimumEthBalance {
		utils.Warnf("[ isWiserNeedBePicked ] wiser: %v's balance is too less: %v", wiser.Address, wiser.EthBalance)
		return false
	}

	if wiser.IsContract > 0 && utils.Contains(config.FamousRouters, wiser.Address) {
		utils.Warnf("[ isWiserNeedBePicked ] wiser: %v is router!", wiser.Address)
		return false
	}

	// 如果没有权重, 也不保存
	if wiser.Weight <= 0 {
		utils.Warnf("[ isWiserNeedBePicked ] wiser weight is too less. wiser: %v", wiser.Address)
		return false
	}

	// 新增要求, 需要有效交易中的trend交易占比大
	if wiser.ValidTrendTradeRatio < w.set.Config.ValidTrendTradeRatio {
		utils.Warnf("[ isWiserNeedBePicked ] wiser ValidTrendTradeRatio is too less. wiser: %v", wiser.Address)
		return false
	}

	return true
}

func (w *Wiser) inspectAccountByDeals(account string, deals []schema.BiDeal) schema.Wiser {
	wiser := schema.Wiser{
		WiserInfo: schema.WiserInfo{
			Address: account,
		},
		CreatedAt: time.Now(),
	}

	winTimes := 0
	totalCost := float64(0)

	frontrunTimes := 0
	rushTimes := 0 // 包含gamble
	trendTimes := 0
	validTrendTimes := 0

	buyMev := 0
	buyFresh := 0
	buySubnew := 0

	buyZeroToken := 0 // 归零币次数

	deflatPairTimes := 0

	for _, deal := range deals {
		wiser.TotalTradeCount++

		// 买入行为分析
		if deal.BuyType == schema.BI_DEAL_BUY_TYPE_MEV {
			buyMev++
		} else if deal.BuyType == schema.BI_DEAL_BUY_TYPE_FRESH {
			buyFresh++
		} else if deal.BuyType == schema.BI_DEAL_BUY_TYPE_SUBNEW {
			buySubnew++
		}

		// 截止当前时间, 是否代币已经归零. 1+yield是因为亏钱为负数
		// 如果 未计算出结果, 此时不计为0
		if (1 + deal.UptoTodayYield) <= w.set.Config.DealDefiniteLoss {
			buyZeroToken++
		}

		// 持有行为分析
		if deal.BiDealType == schema.BI_DEAL_TYPE_ARBI ||
			deal.BiDealType == schema.BI_DEAL_TYPE_FRONTRUN {
			frontrunTimes++
		} else if deal.BiDealType == schema.BI_DEAL_TYPE_TREND {
			trendTimes++
		} else { // gamble 也算rush
			rushTimes++
		}

		if deal.BuyPairHackType == schema.PAIR_MAINTOKEN_HACK_TYPE_DEFLAT || deal.BuyPairHackType == schema.PAIR_MAINTOKEN_HACK_TYPE_SCAM {
			deflatPairTimes++
			continue // 不进行valid统计
		}

		if deal.BuyType == schema.BI_DEAL_BUY_TYPE_MEV || deal.BiDealType == schema.BI_DEAL_TYPE_ARBI || deal.BiDealType == schema.BI_DEAL_TYPE_FRONTRUN || deal.BiDealType == schema.BI_DEAL_TYPE_GAMBLE_TRADE {
			// 同区块内买入或极快卖出不统计胜率, 不进行valid统计
			continue
		}

		// continue后即均为有效交易
		if deal.BiDealType == schema.BI_DEAL_TYPE_TREND {
			validTrendTimes++
		}

		// 分析每一笔valid的deal, 做数据统计
		wiser.ValidTradeCount++
		wiser.TotalWinValue += deal.Earn

		totalCost += math.Abs(deal.Earn / deal.EarnChange)

		if deal.EarnChange >= w.set.Config.DealProfitTarget {
			winTimes++ // 盈利率合格
		}

	}

	if wiser.ValidTradeCount > 0 {
		wiser.WinRatio = float64(winTimes) / float64(wiser.ValidTradeCount)
		wiser.EarnValuePerDeal = wiser.TotalWinValue / float64(wiser.ValidTradeCount)
	}
	if totalCost > 0 {
		wiser.AverageEarnRatio = wiser.TotalWinValue / totalCost
	}

	wiser.TradeCntPerMonth = float64(wiser.ValidTradeCount) * (float64(w.set.Config.LatestSwapSeconds) / 60 / 60 / 24 / 30)

	w.updateWiserTradeInfo(&wiser, frontrunTimes, rushTimes, trendTimes, validTrendTimes)
	w.updateWiserBuyTypeInfo(&wiser, buyMev, buyFresh, buySubnew)
	w.updateOtherProfile(&wiser, deflatPairTimes, buyZeroToken)

	wiser.Epoch = w.epoch
	wiser.AddressWithEpoch = fmt.Sprintf("%v_%v", wiser.Address, wiser.Epoch)

	return wiser
}

func (w *Wiser) updateWiserBuyTypeInfo(wiser *schema.Wiser, buyMev, buyFresh, buySubnew int) {
	if wiser.TotalTradeCount <= 0 {
		return
	}

	wiser.BuyMevRatio = float64(buyMev) / float64(wiser.TotalTradeCount)
	wiser.BuyFreshRatio = float64(buyFresh) / float64(wiser.TotalTradeCount)
	wiser.BuySubnewRatio = float64(buySubnew) / float64(wiser.TotalTradeCount)
}

func (w *Wiser) updateWiserTradeInfo(wiser *schema.Wiser, frontrunTimes, rushTimes, trendTimes, validTrendTimes int) {
	if wiser.TotalTradeCount <= 0 {
		return
	}

	wiser.FrontrunTradeRatio = float64(frontrunTimes) / float64(wiser.TotalTradeCount)
	wiser.RushTradeRatio = float64(rushTimes) / float64(wiser.TotalTradeCount)
	wiser.TrendTradeRatio = float64(trendTimes) / float64(wiser.TotalTradeCount)

	if wiser.ValidTradeCount > 0 {
		wiser.ValidTrendTradeRatio = float64(validTrendTimes) / float64(wiser.ValidTradeCount)
	}
}

func (w *Wiser) updateWiserWeight(wiser *schema.Wiser) {
	// 判断交易频率
	if wiser.TradeCntPerMonth < w.set.Config.DealThresholdPerMon {
		utils.Debugf("[ updateWiserWeight ] TradeCntPerMonth too less: %v", wiser.TradeCntPerMonth)
		return
	}

	// 判断胜率
	if wiser.WinRatio < w.set.Config.WinRatioTarget {
		utils.Debugf("[ updateWiserWeight ] WinRatio too less: %v", wiser.WinRatio)
		return
	}

	// 开始算权重. 胜率占70%, 盈利比例占30%
	weightWinRatio := (wiser.WinRatio - 0.6) / 0.4
	frequency := float64(wiser.ValidTradeCount) / 30.0
	if frequency > 1.0 {
		frequency = 1.0
	}
	weightWinRatio = weightWinRatio * frequency

	weightEarnRatio := wiser.AverageEarnRatio
	if weightEarnRatio > 1 {
		weightEarnRatio = 1
	}
	if weightEarnRatio < -1 {
		weightEarnRatio = -1
	}

	// 权重分配
	wiser.Weight = int(weightWinRatio*80 + weightEarnRatio*20)
}

func (w *Wiser) updateOtherProfile(wiser *schema.Wiser, deflatPairTimes, buyZeroToken int) {
	// 某个账户的交易类型最多，则就是某类型?
	maxValue := wiser.FrontrunTradeRatio
	if wiser.TrendTradeRatio > maxValue {
		maxValue = wiser.TrendTradeRatio
	}
	if wiser.RushTradeRatio > maxValue {
		maxValue = wiser.RushTradeRatio
	}

	if maxValue == wiser.FrontrunTradeRatio {
		wiser.Type = schema.WISER_TRADER_TYPE_FRONTRUN
	} else if maxValue == wiser.RushTradeRatio {
		wiser.Type = schema.WISER_TRADER_TYPE_RUSH
	} else {
		wiser.Type = schema.WISER_TRADER_TYPE_STEADY
	}

	if wiser.TotalTradeCount > 0 {
		wiser.BuyDeflatTokenRatio = float64(deflatPairTimes) / float64(wiser.TotalTradeCount)

		wiser.BuyZeroTokenRatio = float64(buyZeroToken) / float64(wiser.TotalTradeCount)
	}
}

func TEST_WISER() {
	options := &options.FindOptions{}
	options = options.SetSort(bson.D{{Key: "weight", Value: -1}})
	filter := bson.M{}
	filter["epoch"] = "20240127"
	address_oldS, _, _ := wiser.GetWisers(options, &filter, chain.GetMongo().Database("sfilter"))
	address_newS, _, _ := wiser.GetWisers(options, &filter, chain.GetMongo().Database("creat"))

	var old, new []string
	for _, addr := range address_oldS {
		old = append(old, addr.Address)
	}

	for _, addr := range address_newS {
		new = append(new, addr.Address)
	}

	var same, old_have, new_have []string
	for _, addr := range old {
		if utils.Contains(new, addr) {
			// 如果old在new中有, same append
			same = append(same, addr)
		} else {
			old_have = append(old_have, addr)
		}
	}

	for _, addr := range new {
		if !utils.Contains(old, addr) {
			new_have = append(new_have, addr)
		}
	}

	fmt.Println("same len: ", len(same))
	for _, addr := range same {
		fmt.Println(addr)
	}

	fmt.Println("\nold have len: ", len(old_have))
	for _, addr := range old_have {
		fmt.Println(addr)
	}

	fmt.Println("new have len: ", len(new_have))
	for _, addr := range new_have {
		fmt.Println(addr)
	}

}
