package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 定义 Pair map
type PairMap map[string]*Pair

type Pair struct {
	InfoOnChain       `bson:",inline"`
	InfoOnPairCreated `bson:",inline"`
	InfoOnPools       `bson:",inline"`

	TradeInfoForPair `bson:",inline"`

	CacheFields `bson:",inline"`

	UpdatedAt time.Time `json:"-" bson:"updatedAt"`
	CreatedAt time.Time `json:"-" bson:"createdAt"`
}

// 临时字段, 不写db与json
type CacheFields struct {
	ReservedString string  `json:"-" bson:"-"`
	ReservedFloat  float64 `json:"-" bson:"-"`
	ReservedInt    int     `json:"-" bson:"-"`
}

type InfoOnPairCreated struct {
	Type int `json:"type" bson:"type"` // 类型是uniswap v2还是v3等

	PairCreatedBlockNo uint64 `json:"pairCreatedBlockNo" bson:"pairCreatedBlockNo"`
	PairCreatedHash    string `json:"pairCreatedHash" bson:"pairCreatedHash"`

	PairFee int64 `json:"pairFee" bson:"pairFee"`
}

// add liquidity etc..
type InfoOnPools struct {
	// 相当于创建pair的时间了
	FirstAddPoolBlockNo uint64    `json:"firstAddPoolBlockNo" bson:"firstAddPoolBlockNo"`
	FirstAddPoolTime    time.Time `json:"firstAddPoolTime" bson:"firstAddPoolTime"`
	FirstAddTxHash      string    `json:"firstAddTxHash" bson:"firstAddTxHash"`
	FirstAddGasPrice    string    `json:"firstAddGasPrice" bson:"firstAddGasPrice"`
}

type InfoOnChain struct {
	Address string `json:"address" bson:"address"` // 地址

	Token0 string `json:"token0" bson:"token0"` // 代币0
	Token1 string `json:"token1" bson:"token1"` // 代币1

	Decimal0 uint8 `json:"decimal0" bson:"decimal0"`
	Decimal1 uint8 `json:"decimal1" bson:"decimal1"`

	// 池子流动性, 包含屌丝币价值
	LiquidityInUsd float64 `json:"liquidityInUsd" bson:"liquidityInUsd"`
	// 价值币流动性, 即使是 eth/usdt, 也只算一半
	// 这么处理原因是 v3池子可能不是 对等的, 可能屌丝币很多导致失真
	ValueCoinLiquidity float64 `json:"valueCoinLiquidity" bson:"valueCoinLiquidity"`

	Token0UsdValue float64 `json:"token0UsdValue" bson:"token0UsdValue"`
	Token1UsdValue float64 `json:"token1UsdValue" bson:"token1UsdValue"`

	PairName string `json:"pairName" bson:"pairName"` // 如 pepe/weths

	// 是否是通缩币、坑人币等
	MainTokenHackType int `json:"mainTokenHackType" bson:"mainTokenHackType"`
}

const (
	PAIR_MAINTOKEN_HACK_TYPE_UNINIT int = iota

	PAIR_MAINTOKEN_HACK_TYPE_UNKNOWN

	PAIR_MAINTOKEN_HACK_TYPE_NORMAL

	PAIR_MAINTOKEN_HACK_TYPE_DEFLAT // 通缩币
	PAIR_MAINTOKEN_HACK_TYPE_SCAM   // 坑人币

	PAIR_MAINTOKEN_HACK_TYPE_EMPTY_BALANCE // pair没钱了
)

// 额外信息, 尤其与交易相关, 可以用于排序搜索
// 注意搜索排序的时候要带上 UpdateAt 信息
type TradeInfoForPair struct {
	// 1h 内tx数, 用于排序等
	TxNumIn1h int `json:"txNumIn1h" bson:"txNumIn1h"`

	// 24h 内tx数
	TxNumIn24h int `json:"txNumIn24h" bson:"txNumIn24h"`

	// 最近1h相对于上一个小时, tx的增减比例
	TxNumChangeIn1h float32 `json:"txNumChangeIn1h" bson:"txNumChangeIn1h"`

	// 最近1d相对于前一天, tx的增减比例
	TxNumChangeIn24h float32 `json:"txNumChangeIn24h" bson:"txNumChangeIn24h"`

	// 1h 价格变化
	PriceChangeIn1h float32 `json:"priceChangeIn1h" bson:"priceChangeIn1h"`

	// 24h 价格变化
	PriceChangeIn24h float32 `json:"priceChangeIn24h" bson:"priceChangeIn24h"`

	// 直接float存储, 毕竟这里只是展示, 不需要高精度计算
	Price      float64 `json:"price" bson:"price"`
	PriceInUsd float64 `json:"priceInUsd" bson:"priceInUsd"`

	// 1h内的交易金额(usd); float64足够存储, 毕竟 1亿 就已经很大了
	VolumeByUsdIn1h float64 `json:"volumeByUsdIn1h" bson:"volumeByUsdIn1h"`

	// 24h内的交易金额(usd)
	VolumeByUsdIn24h float64 `json:"volumeByUsdIn24h" bson:"volumeByUsdIn24h"`

	// tradeinfo更新时间, 用于获取最新有交易的pair
	TradeInfoUpdatedAt time.Time `json:"tradeInfoUpdatedAt" bson:"tradeInfoUpdatedAt"`
}

var PairIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "updatedAt", Value: -1}},
		Options: options.Index().SetName("updatedAt_index"),
	},
	{
		Keys:    bson.D{{Key: "address", Value: 1}},
		Options: options.Index().SetName("address_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "token0", Value: 1}},
		Options: options.Index().SetName("token0_index"),
	},
	{
		Keys:    bson.D{{Key: "token1", Value: 1}},
		Options: options.Index().SetName("token1_index"),
	},
	{
		Keys:    bson.D{{Key: "pairCreatedBlockNo", Value: 1}},
		Options: options.Index().SetName("pairCreatedBlockNo_index").SetSparse(true),
	},
	{
		Keys:    bson.D{{Key: "firstAddPoolBlockNo", Value: 1}},
		Options: options.Index().SetName("firstAddPoolBlockNo_index").SetSparse(true),
	},
	{
		Keys:    bson.D{{Key: "firstAddPoolTime", Value: 1}},
		Options: options.Index().SetName("firstAddPoolTime_index").SetSparse(true),
	},
	{
		Keys:    bson.D{{Key: "pairName", Value: 1}},
		Options: options.Index().SetName("pairName_index").SetSparse(true),
	},
	{
		Keys:    bson.D{{Key: "liquidityInUsd", Value: 1}},
		Options: options.Index().SetName("liquidityInUsd_index").SetSparse(true),
	},
	{
		Keys:    bson.D{{Key: "valueCoinLiquidity", Value: 1}},
		Options: options.Index().SetName("valueCoinLiquidity_index").SetSparse(true),
	},

	//  数值加索引方便排序？
	{
		Keys:    bson.D{{Key: "txNumIn1h", Value: 1}},
		Options: options.Index().SetName("txNumIn1h_index"),
	},
	{
		Keys:    bson.D{{Key: "txNumIn24h", Value: 1}},
		Options: options.Index().SetName("txNumIn24h_index"),
	},
	{
		Keys:    bson.D{{Key: "txNumChangeIn1h", Value: 1}},
		Options: options.Index().SetName("txNumChangeIn1h_index"),
	},
	{
		Keys:    bson.D{{Key: "txNumChangeIn24h", Value: 1}},
		Options: options.Index().SetName("txNumChangeIn24h_index"),
	},
	{
		Keys:    bson.D{{Key: "priceChangeIn1h", Value: 1}},
		Options: options.Index().SetName("priceChangeIn1h_index"),
	},

	{
		Keys:    bson.D{{Key: "priceChangeIn24h", Value: 1}},
		Options: options.Index().SetName("priceChangeIn24h_index"),
	},
	{
		Keys:    bson.D{{Key: "price", Value: 1}},
		Options: options.Index().SetName("price_index"),
	},
	{
		Keys:    bson.D{{Key: "volumeByUsdIn1h", Value: 1}},
		Options: options.Index().SetName("volumeByUsdIn1h_index"),
	},
	{
		Keys:    bson.D{{Key: "volumeByUsdIn24h", Value: 1}},
		Options: options.Index().SetName("volumeByUsdIn24h_index"),
	},
	{
		Keys:    bson.D{{Key: "tradeInfoUpdatedAt", Value: 1}},
		Options: options.Index().SetName("tradeInfoUpdatedAt_index"),
	},
}
