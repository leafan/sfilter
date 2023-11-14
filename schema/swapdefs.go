package schema

import (
	"math/big"
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Swap struct {
	BlockNo         uint64  // 区块号
	TxHash          string  // tx hash
	Position        uint    // tx 在本区块中的序号
	PairAddr        string  // pair addr
	CurrentEthPrice float64 // 当前区块(时间)的eth价格in usd

	GasPrice string // 使用的gas price
	GasInEth string // 消耗的gas值

	Token0 string
	Token1 string

	// 买卖的token
	// 如果都是屌丝币, 以Token0为主token, token1为quote币
	// 如果都是quote币, 则当屌丝币分析
	MainToken int // 0表示token0; 1表示token1
	Direction int // 0买; 1卖

	Operator      string // msg.sender
	OperatorNonce uint64 // msg.sender 当时的nonce

	Sender    string // event中记录的sender
	Recipient string // event中记录的sender

	Price             *big.Int // 买卖价格, 以 mainToken(/decimal) / quoteToken(/decimal) * 1e18
	AmountOfMainToken *big.Int // main token amount / decimal

	LogIndexWithTx string // tx hash 以及 log 在本区块中的序号，以作为唯一标识
	CreatedAt      time.Time
}

var SwapIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdat", Value: -1}},
		Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(config.SwapSaveTime),
	},
	{
		Keys:    bson.D{{Key: "txhash", Value: 1}},
		Options: options.Index().SetName("txhash_index"),
	},
	{
		Keys:    bson.D{{Key: "logindexwithtx", Value: 1}},
		Options: options.Index().SetName("logindexwithtx_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "receiver", Value: 1}},
		Options: options.Index().SetName("receiver_index"),
	},
}
