package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// db中的实时数据存储
// config key: wiser_config
type WiserDBConfig struct {
	Epoch  string `json:"epoch" bson:"epoch"`   //最新的epoch
	Config string `json:"config" bson:"config"` //最新的epoch

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// deal 持有类型
const (
	BI_DEAL_TYPE_UNKNOWN int = iota

	// arbitrage 交易机器, 定义为同区块进出
	// 或者是 usdt->eth->xxx, 导致该deal买入eth又卖出eth
	// 不应该被统计到合理deal里面, 一般都是 usdt->eth 价值币之间的转换
	BI_DEAL_TYPE_ARBI

	// frontrun夹子机器人, 定义为3-5(由配置确定)以内的区块卖出
	BI_DEAL_TYPE_FRONTRUN

	// 超高频投机交易, 定义为5min之内的进出
	BI_DEAL_TYPE_GAMBLE_TRADE

	// 高频交易, 定义为30min之内的进出
	BI_DEAL_TYPE_RUSH_TRADE

	BI_DEAL_TYPE_TREND //  正常趋势交易
)

// deal 买入风格类型
const (
	BI_DEAL_BUY_TYPE_UNKNOWN int = iota

	BI_DEAL_BUY_TYPE_MEV    // 创建pair的同区块就买入
	BI_DEAL_BUY_TYPE_FRESH  // 创建pair后5min内买入
	BI_DEAL_BUY_TYPE_SUBNEW // 次新币, 创建pair后7天内买入
	BI_DEAL_BUY_TYPE_TREND  // 正常买卖
)

// wiser 交易风格类型
const (
	WISER_TRADER_TYPE_UNKNOWN int = iota

	WISER_TRADER_TYPE_FRONTRUN
	WISER_TRADER_TYPE_RUSH
	WISER_TRADER_TYPE_STEADY
)

// wiser 交易类型
const (
	TRADE_TYPE_UNKNOWN int = iota

	TRADE_TYPE_SWAP     // swap
	TRADE_TYPE_TRANSFER // transfer

	TRADE_TYPE_LIQUIDATION // 强制结算, 如用户没卖但是清零了
)

// 优秀地址定义
type Wiser struct {
	WiserInfo `bson:",inline"` // wiser本身定义

	DealDetail `bson:",inline"` // 统计数据记录

	// 因为会定期计算与更新wiser, 因此记录下当前运算epoch, 方便对比
	// 暂时不用
	Epoch string `json:"epoch" bson:"epoch"`

	AddressWithEpoch string `json:"addressWithEpoch" bson:"addressWithEpoch"` // unique key

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type WiserInfo struct {
	Address string `json:"address" bson:"address"` // 地址

	Weight int `json:"weight" bson:"weight"` // 计算出的权重

	EthBalance float64 `json:"ethBalance" bson:"ethBalance"`

	IsContract int `json:"isContract" bson:"isContract"`

	// 交易风格, 如快进快出等
	Type int `json:"type" bson:"type"`
}

// 统计数据详情
type DealDetail struct {
	WinRatio float64 `json:"winRatio" bson:"winRatio"` // 盈利比例

	TotalTradeCount int `json:"totalTradeCount" bson:"totalTradeCount"` // 有效交易总笔数
	ValidTradeCount int `json:"validTradeCount" bson:"validTradeCount"` // 有效交易总笔数

	// 统计下arbi、frontrun、trend 等交易比例
	FrontrunTradeRatio float64 `json:"frontrunTradeRatio" bson:"frontrunTradeRatio"`
	RushTradeRatio     float64 `json:"rushTradeRatio" bson:"rushTradeRatio"`

	// trend交易占整体交易比例
	TrendTradeRatio float64 `json:"trendTradeRatio" bson:"trendTradeRatio"`
	// 有效trend交易占有效交易比例
	ValidTrendTradeRatio float64 `json:"validTrendTradeRatio" bson:"validTrendTradeRatio"`

	// 统计下Buyer的一些比例数据
	BuyMevRatio    float64 `json:"buyMevRatio" bson:"buyMevRatio"`
	BuyFreshRatio  float64 `json:"buyFreshRatio" bson:"buyFreshRatio"`
	BuySubnewRatio float64 `json:"buySubnewRatio" bson:"buySubnewRatio"`

	// 购买通缩币或坑人币比例
	BuyDeflatTokenRatio float64 `json:"buyDeflatTokenRatio" bson:"buyDeflatTokenRatio"`

	// 购买截止至今的归零币比例
	BuyZeroTokenRatio float64 `json:"buyZeroTokenRatio" bson:"buyZeroTokenRatio"`

	// 每月平均交易次数, 算法从第一笔买到最后一笔卖算周期时长
	TradeCntPerMonth float64 `json:"tradeCntPerMonth" bson:"tradeCntPerMonth"`

	TotalWinValue    float64 `json:"totalWinValue" bson:"totalWinValue"`       // 盈利总金额
	EarnValuePerDeal float64 `json:"earnValuePerDeal" bson:"earnValuePerDeal"` // 平均每笔盈利

	AverageEarnRatio float64 `json:"averageEarnRatio" bson:"averageEarnRatio"` // 平均盈利比例
}

// 一笔买卖的定义, bi 包含双向的意思
type BiDeal struct {
	Account   string `json:"account" bson:"account"`     // 用户地址
	Token     string `json:"token" bson:"token"`         // token地址
	TokenName string `json:"tokenName" bson:"tokenName"` // 方便人阅读

	// buy
	BuyTxHash   string `json:"buyTxHash" bson:"buyTxHash"` // 第一笔买入tx
	BuyBlockNo  uint64 `json:"buyBlockNo" bson:"buyBlockNo"`
	BuyPair     string `json:"buyPair" bson:"buyPair"` // 买入的pair
	BuyPairType int    `json:"-" bson:"-"`             // 临时变量, 用于判断是v2还是v3

	BuyPairAge int `json:"buyPairAge" bson:"buyPairAge"` // 买入时该pair的创建时长
	BuyType    int `json:"buyType" bson:"buyType"`       // 买入风格, 如迅速买入等

	// Token是否是通缩币、坑人币等
	BuyPairHackType int `json:"buyPairHackType" bson:"buyPairHackType"`

	BuyValue  float64 `json:"buyValue" bson:"buyValue"`
	BuyAmount float64 `json:"buyAmount" bson:"buyAmount"`
	BuyPrice  float64 `json:"buyPrice" bson:"buyPrice"`

	// sell
	// 第一笔卖出tx. 由于一定只有一笔sell, 因此可以用sell tx做为deal的 unique key
	// 由于可能出现一笔tx里面卖出多笔token, 所以需要 sellTxHash_Token 为key
	SellTxHashWithToken string `json:"sellTxHashWithToken" bson:"sellTxHashWithToken"`
	SellBlockNo         uint64 `json:"sellBlockNo" bson:"sellBlockNo"`
	SellPair            string `json:"sellPair" bson:"sellPair"`

	SellTime time.Time `json:"sellTime" bson:"sellTime"`

	SellValue  float64 `json:"sellValue" bson:"sellValue"`
	SellAmount float64 `json:"sellAmount" bson:"sellAmount"`
	SellPrice  float64 `json:"sellPrice" bson:"sellPrice"`
	SellType   int     `json:"sellType" bson:"sellType"`

	// summary
	Earn       float64 `json:"earn" bson:"earn"`             // 盈利金额
	EarnChange float64 `json:"earnChange" bson:"earnChange"` //盈利比例
	HoldBlocks uint64  `json:"holdBlocks" bson:"holdBlocks"` // 持有的区块数

	// 如果持有至今的盈利率, 防止坑人币最终都归零
	UptoTodayYield float64 `json:"uptoTodayYield" bson:"uptoTodayYield"`

	// 持有风格类型
	BiDealType int `json:"biDealType" bson:"biDealType"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// pair_addr为key, value为rank数组, 以createdAt倒排
type HRankMap map[string][]HotPairRank

// 排名列表保存, 每分钟保存前50名
// 保存一个月有 60*24*30*50 = 216w
type HotPairRank struct {
	// year_month_day or year_month_day_hour
	PeriodKey string `json:"periodKey" bson:"periodKey"`

	// 被选中买点的时候, 其 tx 排序多少
	SortRank int `json:"sortRank" bson:"sortRank"`

	// unique key, 由 year_month_day_pairAddr or year_month_day_hour_pairAddr 组成
	PeriodKeyWithPair string `json:"periodKeyWithPair" bson:"periodKeyWithPair"`

	MainToken     string        `json:"mainToken" bson:"mainToken"`
	PairAddress   string        `json:"pairAddress" bson:"pairAddress"`
	PairName      string        `json:"pairName" bson:"pairName"`
	PairLiquidity float64       `json:"pairLiquidity" bson:"pairLiquidity"`
	PairAge       time.Duration `json:"pairAge" bson:"pairAge"`

	// 把pair中的对应排名相关信息全拷过来
	TradeInfoForPair `bson:",inline"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// 买卖策略记录
type BiTrade struct {
	MainToken     string        `json:"mainToken" bson:"mainToken"`
	PairAddress   string        `json:"pairAddress" bson:"pairAddress"`
	PairName      string        `json:"pairName" bson:"pairName"`
	PairLiquidity float64       `json:"pairLiquidity" bson:"pairLiquidity"`
	PairAge       time.Duration `json:"pairAge" bson:"pairAge"`

	// buy info
	BuyPrice  float64   `json:"buyPrice" bson:"buyPrice"`
	BuyTime   time.Time `json:"buyTime" bson:"buyTime"`
	BuyReason string    `json:"buyReason" bson:"buyReason"`
	SortRank  int       `json:"sortRank" bson:"sortRank"` // 被选中买点的时候, 其 tx1h 排序多少
	TxNumIn1h int       `json:"txNumIn1h" bson:"txNumIn1h"`

	// sell info
	SellPrice  float64   `json:"sellPrice" bson:"sellPrice"`
	SellTime   time.Time `json:"sellTime" bson:"sellTime"`
	SellReason string    `json:"sellReason" bson:"sellReason"` // 卖出原因, 方便查看分析

	// summarize
	HighestPrice float64       `json:"highestPrice" bson:"highestPrice"` // 周期中最高价格
	EarnRatio    float64       `json:"earnRatio" bson:"earnRatio"`
	HoldTime     time.Duration `json:"holdTime" bson:"holdTime"`

	// 0表示已买入尚未卖出, 1表示已卖出, 可以重新买入
	Status int `json:"status" bson:"status"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// key是token, 数组为token内的所有trade
type AccountTrades map[string][]AccountTokenTrade

// 一个地址的某一个token，经过处理后的交易记录
// 不存db, 用于内存中计算
type AccountTokenTrade struct {
	BlockNo  uint64
	TxHash   string
	Position uint

	TradeTime time.Time

	Pair     string // pair address
	PairType int

	Type      int
	Direction int

	Amount     float64
	USDValue   float64 // 法币价值
	PriceInUSD float64 // 法币价格
}

var HotPairRankIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "periodKeyWithPair", Value: 1}},
		Options: options.Index().SetName("periodKeyWithPair_index").SetUnique(true),
	},

	{
		Keys:    bson.D{{Key: "mainToken", Value: -1}},
		Options: options.Index().SetName("mainToken_index"),
	},
	{
		Keys:    bson.D{{Key: "pairAddress", Value: -1}},
		Options: options.Index().SetName("pairAddress_index"),
	},
	{
		Keys:    bson.D{{Key: "pairAge", Value: -1}},
		Options: options.Index().SetName("pairAge_index"),
	},
	{
		Keys:    bson.D{{Key: "sortRank", Value: -1}},
		Options: options.Index().SetName("sortRank_index"),
	},

	{
		Keys:    bson.D{{Key: "pairLiquidity", Value: -1}},
		Options: options.Index().SetName("pairLiquidity_index"),
	},
}

var BiTradeIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},

	{
		Keys:    bson.D{{Key: "mainToken", Value: -1}},
		Options: options.Index().SetName("mainToken_index"),
	},
	{
		Keys:    bson.D{{Key: "pairAddress", Value: -1}},
		Options: options.Index().SetName("pairAddress_index"),
	},
	{
		Keys:    bson.D{{Key: "earnRatio", Value: -1}},
		Options: options.Index().SetName("earnRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "status", Value: -1}},
		Options: options.Index().SetName("status_index"),
	},
}

var WiserIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		// 以sell tx 为 unique key
		Keys:    bson.D{{Key: "addressWithEpoch", Value: 1}},
		Options: options.Index().SetName("addressWithEpoch_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index"),
	},

	{
		Keys:    bson.D{{Key: "weight", Value: 1}},
		Options: options.Index().SetName("weight_index"),
	},

	{
		Keys:    bson.D{{Key: "winRatio", Value: 1}},
		Options: options.Index().SetName("winRatio_index"),
	},

	{
		Keys:    bson.D{{Key: "validTradeCount", Value: 1}},
		Options: options.Index().SetName("validTradeCount_index"),
	},
	{
		Keys:    bson.D{{Key: "totalTradeCount", Value: 1}},
		Options: options.Index().SetName("totalTradeCount_index"),
	},

	{
		Keys:    bson.D{{Key: "buyZeroTokenRatio", Value: 1}},
		Options: options.Index().SetName("buyZeroTokenRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "buyMevRatio", Value: 1}},
		Options: options.Index().SetName("buyMevRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "buyFreshRatio", Value: 1}},
		Options: options.Index().SetName("buyFreshRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "buySubnewRatio", Value: 1}},
		Options: options.Index().SetName("buySubnewRatio_index"),
	},

	{
		Keys:    bson.D{{Key: "tradeCntPerMonth", Value: 1}},
		Options: options.Index().SetName("tradeCntPerMonth_index"),
	},
	{
		Keys:    bson.D{{Key: "totalWinValue", Value: 1}},
		Options: options.Index().SetName("totalWinValue_index"),
	},
	{
		Keys:    bson.D{{Key: "AverageEarnRatio", Value: 1}},
		Options: options.Index().SetName("AverageEarnRatio_index"),
	},

	{
		Keys:    bson.D{{Key: "type", Value: 1}},
		Options: options.Index().SetName("type_index"),
	},
	{
		Keys:    bson.D{{Key: "frontrunTradeRatio", Value: 1}},
		Options: options.Index().SetName("frontrunTradeRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "rushTradeRatio", Value: 1}},
		Options: options.Index().SetName("rushTradeRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "trendTradeRatio", Value: 1}},
		Options: options.Index().SetName("trendTradeRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "buyPairHackType", Value: 1}},
		Options: options.Index().SetName("buyPairHackType_index"),
	},

	{
		Keys:    bson.D{{Key: "epoch", Value: 1}},
		Options: options.Index().SetName("epoch_index"),
	},
}

var BiDealIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		// 以sell tx 为 unique key
		Keys:    bson.D{{Key: "sellTxHashWithToken", Value: 1}},
		Options: options.Index().SetName("sellTxHashWithToken_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "account", Value: 1}},
		Options: options.Index().SetName("account_index"),
	},
	{
		Keys:    bson.D{{Key: "token", Value: 1}},
		Options: options.Index().SetName("token_index"),
	},
	{
		Keys:    bson.D{{Key: "buyPrice", Value: 1}},
		Options: options.Index().SetName("buyPrice_index"),
	},
	{
		Keys:    bson.D{{Key: "sellPrice", Value: 1}},
		Options: options.Index().SetName("sellPrice_index"),
	},
	{
		Keys:    bson.D{{Key: "sellType", Value: 1}},
		Options: options.Index().SetName("sellType_index"),
	},
	{
		Keys:    bson.D{{Key: "sellTime", Value: 1}},
		Options: options.Index().SetName("sellTime_index"),
	},

	{
		Keys:    bson.D{{Key: "buyType", Value: 1}},
		Options: options.Index().SetName("buyType_index"),
	},
	{
		Keys:    bson.D{{Key: "earn", Value: 1}},
		Options: options.Index().SetName("earn_index"),
	},
	{
		Keys:    bson.D{{Key: "earnChange", Value: 1}},
		Options: options.Index().SetName("earnChange_index"),
	},
	{
		Keys:    bson.D{{Key: "biDealType", Value: 1}},
		Options: options.Index().SetName("biDealType_index"),
	},
	{
		Keys:    bson.D{{Key: "buyDeflatTokenRatio", Value: 1}},
		Options: options.Index().SetName("buyDeflatTokenRatio_index"),
	},
	{
		Keys:    bson.D{{Key: "holdBlocks", Value: 1}},
		Options: options.Index().SetName("holdBlocks_index"),
	},
}
