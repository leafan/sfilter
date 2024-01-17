package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TokenList []string

// 活跃账户, key为账户地址, value为交易过的token list
type ActiveAccount map[string]TokenList

const (
	BI_DEAL_TYPE_UNKNOWN int = iota

	BI_DEAL_TYPE_CLASSIC //  正常趋势交易

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
)

const (
	BI_DEAL_STATUS_UNINIT int = iota

	BI_DEAL_STATUS_BUY  // 买入阶段
	BI_DEAL_STATUS_SELL // 卖出阶段

	BI_DEAL_STATUS_FINISHED // 结束
)

// 优秀地址定义
type Wiser struct {
	Address string `json:"address" bson:"address"` // 地址
	Weight  int    `json:"weight" bson:"weight"`   // 计算出的权重

	// 因为会定期计算与更新wiser, 因此记录下当前运算epoch
	// 方便对比，或累计运算权重
	Epoch string `json:"epoch" bson:"epoch"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// 一笔买卖的定义, bi 包含双向的意思
type BiDeal struct {
	Account   string `json:"account" bson:"account"`     // 用户地址
	Token     string `json:"token" bson:"token"`         // token地址
	TokenName string `json:"tokenName" bson:"tokenName"` // 方便人阅读

	// buy
	BuyTxHash  string `json:"buyTxHash" bson:"buyTxHash"` // 第一笔买入tx
	BuyBlockNo uint64 `json:"buyBlockNo" bson:"buyBlockNo"`

	BuyValue  float64 `json:"buyValue" bson:"buyValue"`
	BuyAmount float64 `json:"buyAmount" bson:"buyAmount"`
	BuyPrice  float64 `json:"buyPrice" bson:"buyPrice"`

	// sell
	// 第一笔卖出tx. 由于一定只有一笔sell, 因此可以用sell tx做为deal的 unique key
	// 由于可能出现一笔tx里面卖出多笔token, 所以需要 sellTxHash_Token 为key
	SellTxHashWithToken string `json:"sellTxHashWithToken" bson:"sellTxHashWithToken"`
	SellBlockNo         uint64 `json:"sellBlockNo" bson:"sellBlockNo"`

	SellValue  float64 `json:"sellValue" bson:"sellValue"`
	SellAmount float64 `json:"sellAmount" bson:"sellAmount"`
	SellPrice  float64 `json:"sellPrice" bson:"sellPrice"`
	SellType   int     `json:"sellType" bson:"sellType"`

	// summary
	Earn       float64 `json:"earn" bson:"earn"`             // 盈利金额
	EarnChange float64 `json:"earnChange" bson:"earnChange"` //盈利比例
	HoldBlocks uint64  `json:"holdBlocks" bson:"holdBlocks"` // 持有的区块数

	BiDealType int `json:"biDealType" bson:"biDealType"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// wiser 定义的一些常量, 只用于wiser模块
const (
	WISER_TYPE_UNKNOWN int = iota

	WISER_TYPE_SWAP
	WISER_TYPE_TRANSFER
)

// 一个地址的某一个token，经过处理后的交易记录
// 不存db, 用于内存中计算
type AccountTokenTrade struct {
	BlockNo  uint64
	TxHash   string
	Position uint

	Type      int
	Direction int

	Amount     float64
	USDValue   float64 // 法币价值
	PriceInUSD float64 // 法币价格
}

var WiserIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
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
		Keys:    bson.D{{Key: "holdBlocks", Value: 1}},
		Options: options.Index().SetName("holdBlocks_index"),
	},
}
