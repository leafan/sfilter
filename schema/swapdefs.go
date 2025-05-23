package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	SWAP_EVENT_UNKNOWN int = iota

	SWAP_EVENT_RESERVED_1
	SWAP_EVENT_UNISWAPV2_LIKE // uniswapv2 like
	SWAP_EVENT_UNISWAPV3_LIKE // uniswapv3 like

)

const (
	DIRECTION_UNKNOWN int = iota

	DIRECTION_BUY_OR_ADD       // 买 或者 加流动性 等
	DIRECTION_SELL_OR_DECREASE // 卖 或者 减流动性等

	DIRECTION_ARBI            // arbi bot type
	DIRECTION_TOO_COMPLICATED // 一笔log数过复杂不处理
)

type Swap struct {
	BlockNo  uint64 `json:"blockNo" bson:"blockNo"`   // 区块号
	TxHash   string `json:"txHash" bson:"txHash"`     // 交易哈希
	Position uint   `json:"position" bson:"position"` // 交易在本区块中的序号

	PairAddr string `json:"pairAddr" bson:"pairAddr"` // 交易对地址
	PairName string `json:"pairName" bson:"pairName"`
	SwapType int    `json:"swapType" bson:"swapType"` // 交易类型

	CurrentEthPrice float64 `json:"currentEthPrice" bson:"currentEthPrice"` // 当前区块（时间）的eth价格in usd

	GasPrice string `json:"gasPrice" bson:"gasPrice"` // 使用的gas price
	GasInEth string `json:"gasInEth" bson:"gasInEth"` // 消耗的gas值

	Token0 string `json:"token0" bson:"token0"` // 代币0
	Token1 string `json:"token1" bson:"token1"` // 代币1

	MainToken string `json:"mainToken" bson:"mainToken"` // 主token的地址，方便索引
	Direction int    `json:"direction" bson:"direction"` // 买卖方向

	Operator      string `json:"operator" bson:"operator"`           // msg.sender
	OperatorNonce uint64 `json:"operatorNonce" bson:"operatorNonce"` // 操作者nonce

	Trader string `json:"trader" bson:"trader"` // 交易者

	Sender     string  `json:"sender" bson:"sender"`       // 发送者
	Recipient  string  `json:"recipient" bson:"recipient"` // 接收者
	Price      float64 `json:"price" bson:"price"`         // 买卖价格, 以 mainToken(/decimal) / quoteToken(/decimal) * 1e18
	PriceInUsd float64 `json:"priceInUsd" bson:"priceInUsd"`

	AmountOfMainToken float64 `json:"amountOfMainToken" bson:"amountOfMainToken"` // 主代币数量

	AmountOfMainBig  string `json:"-" bson:"amountOfMainBig"`
	AmountOfQuoteBig string `json:"-" bson:"amountOfQuoteBig"`

	VolumeInUsd float64 ` json:"volumeInUsd" bson:"volumeInUsd"` // 本次交易以Usd计价金额

	LogIndexWithTx string `json:"-" bson:"logIndexWithTx"` // tx hash 以及 log 在本区块中的序号，以作为唯一标识

	LogNumInHash int `json:"logNumInHash" bson:"logNumInHash"`

	SwapTime time.Time `json:"swapTime" bson:"swapTime"`

	SwapOmitFields `bson:",inline"`

	CreatedAt time.Time `json:"-" bson:"createdAt"` // 创建时间
}

// 临时变量, 不存入db
type SwapOmitFields struct {
	Decimal0 uint8 `json:"-" bson:"-"`
	Decimal1 uint8 `json:"-" bson:"-"`
}

var SwapIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "swapTime", Value: -1}},
		Options: options.Index().SetName("swapTime_index").SetExpireAfterSeconds(config.SwapSaveTime),
	},
	{
		Keys:    bson.D{{Key: "blockNo", Value: 1}},
		Options: options.Index().SetName("blockNo_index"),
	},
	{
		Keys:    bson.D{{Key: "txHash", Value: 1}},
		Options: options.Index().SetName("txHash_index"),
	},
	{
		Keys:    bson.D{{Key: "position", Value: 1}},
		Options: options.Index().SetName("position_index"),
	},
	{
		Keys:    bson.D{{Key: "logIndexWithTx", Value: 1}},
		Options: options.Index().SetName("logIndexWithTx_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "pairAddr", Value: 1}},
		Options: options.Index().SetName("pairAddr_index"),
	},
	{
		Keys:    bson.D{{Key: "operator", Value: 1}},
		Options: options.Index().SetName("operator_index"),
	},
	{
		Keys:    bson.D{{Key: "trader", Value: 1}},
		Options: options.Index().SetName("trader_index").SetSparse(true),
	},
	{
		Keys:    bson.D{{Key: "mainToken", Value: 1}},
		Options: options.Index().SetName("mainToken_index"),
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
		Keys:    bson.D{{Key: "volumeInUsd", Value: 1}},
		Options: options.Index().SetName("volumeInUsd_index"),
	},
	{
		Keys:    bson.D{{Key: "amountOfMainToken", Value: 1}},
		Options: options.Index().SetName("amountOfMainToken_index"),
	},
}
