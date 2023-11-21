package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Pair struct {
	InfoOnChain `bson:",inline"`

	TradeInfoForPair `bson:",inline"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"` // 创建时间
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"` // 创建时间
}

type InfoOnChain struct {
	Address  string `json:"address" bson:"address"`   // 地址
	Token0   string `json:"token0" bson:"token0"`     // 代币0
	Token1   string `json:"token1" bson:"token1"`     // 代币1
	PairName string `json:"pairName" bson:"pairName"` // 如 pepe/weths

	Type int `json:"type" bson:"type"` // 类型是uniswap v2还是v3等

	PairCreatedBlockNo uint64 `json:"pairCreatedBlockNo" bson:"pairCreatedBlockNo"`
	PairCreatedTime    int64  `json:"pairCreatedTime" bson:"pairCreatedTime"`
}

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
	Price float64 `json:"price" bson:"price"`

	// 1h内的交易金额(usd); float64足够存储, 毕竟 1亿 就已经很大了
	VolumeByUsdIn1h float64 `json:"volumeByUsdIn1h" bson:"volumeByUsdIn1h"`

	// 24h内的交易金额(usd)
	VolumeByUsdIn24h float64 `json:"volumeByUsdIn24h" bson:"volumeByUsdIn24h"`
}

var PairIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
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
		Options: options.Index().SetName("pairCreatedBlockNo_index"),
	},
	{
		Keys:    bson.D{{Key: "pairName", Value: 1}},
		Options: options.Index().SetName("pairName_index"),
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
}
