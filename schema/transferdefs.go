package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 定义一个map, 以tx_token为key, value为该tx内该token的交易list
// 目的: 用于swap的时候谁是真正的token受益方(操作方)
type TxTokenTransfersMap map[string][]*Transfer

type Transfer struct {
	Token string `json:"token" bson:"token"` // 地址

	From string `json:"from" bson:"from"`
	To   string `json:"to" bson:"to"`

	Amount      string `json:"amount" bson:"amount"`
	AmountInUsd string `json:"amountInUsd" bson:"amountInUsd"`

	BlockNo  uint64 `json:"blockNo" bson:"blockNo"`   // 区块号
	TxHash   string `json:"txHash" bson:"txHash"`     // 交易哈希
	Position uint   `json:"position" bson:"position"` // 交易在本区块中的序号

	LogIndexWithTx string `json:"logIndexWithTx" bson:"logIndexWithTx"` // tx hash 以及 log 在本区块中的序号，以作为唯一标识

	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // transfer时间

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

var TransferIndexModel = []mongo.IndexModel{
	{
		// 以他实际timestamp来做丢弃, 防止回溯的时候过多transfer日志
		Keys:    bson.D{{Key: "timestamp", Value: -1}},
		Options: options.Index().SetName("timestamp_index").SetExpireAfterSeconds(config.TransferTableSavetime),
	},
	{
		Keys:    bson.D{{Key: "logIndexWithTx", Value: 1}},
		Options: options.Index().SetName("logIndexWithTx_index").SetUnique(true),
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
		Keys:    bson.D{{Key: "token", Value: 1}},
		Options: options.Index().SetName("token_index"),
	},
	{
		Keys:    bson.D{{Key: "from", Value: 1}},
		Options: options.Index().SetName("from_index"),
	},
	{
		Keys:    bson.D{{Key: "to", Value: 1}},
		Options: options.Index().SetName("to_index"),
	},
}
