package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Swap struct {
	BlockNo  uint64 `json:"blockNo" bson:"blockNo"`   // 区块号
	TxHash   string `json:"txHash" bson:"txHash"`     // 交易哈希
	Position uint   `json:"position" bson:"position"` // 交易在本区块中的序号

	PairAddr string `json:"pairAddr" bson:"pairAddr"` // 交易对地址
	SwapType int    `json:"swapType" bson:"swapType"` // 交易类型,

	CurrentEthPrice float64 `json:"currentEthPrice" bson:"currentEthPrice"` // 当前区块（时间）的eth价格in usd

	GasPrice string `json:"gasPrice" bson:"gasPrice"` // 使用的gas price
	GasInEth string `json:"gasInEth" bson:"gasInEth"` // 消耗的gas值

	Token0    string `json:"token0" bson:"token0"`       // 代币0
	Token1    string `json:"token1" bson:"token1"`       // 代币1
	MainToken string `json:"mainToken" bson:"mainToken"` // 主token的地址，方便索引
	Direction int    `json:"direction" bson:"direction"` // 买卖方向，0表示买，1表示卖

	Operator      string `json:"operator" bson:"operator"`           // msg.sender, 默认当成买卖操作者
	OperatorNonce uint64 `json:"operatorNonce" bson:"operatorNonce"` // 操作者nonce

	Sender            string `json:"sender" bson:"sender"`                       // 发送者
	Recipient         string `json:"recipient" bson:"recipient"`                 // 接收者
	Price             string `json:"price" bson:"price"`                         // 买卖价格, 以 mainToken(/decimal) / quoteToken(/decimal) * 1e18
	AmountOfMainToken string `json:"amountOfMainToken" bson:"amountOfMainToken"` // 主代币数量

	LogIndexWithTx string `json:"logIndexWithTx" bson:"logIndexWithTx"` // tx hash 以及 log 在本区块中的序号，以作为唯一标识

	SwapTime time.Time `json:"swapTime" bson:"swapTime"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"` // 创建时间
}

var SwapIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index").SetExpireAfterSeconds(config.SwapSaveTime),
	},
	{
		Keys:    bson.D{{Key: "blockNo", Value: 1}},
		Options: options.Index().SetName("blockNo_index"),
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
		Keys:    bson.D{{Key: "mainToken", Value: 1}},
		Options: options.Index().SetName("mainToken_index"),
	},
}
