package schema

import (
	"math/big"
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TRANSFER_EVENT_UNKNOWN int = iota

	TRANSFER_EVENT_TRANSFER // 正常的transfer
	TRANSFER_EVENT_SWAP     // swap性质的transfer

)

type Transfer struct {
	Token string `json:"token" bson:"token"` // 地址

	TokenSymbol string `json:"tokenSymbol" bson:"tokenSymbol"`

	From string `json:"from" bson:"from"`
	To   string `json:"to" bson:"to"`

	Amount       float64  `json:"amount" bson:"amount"`
	AmountBigInt *big.Int `json:"-" bson:"-"` // 暂时保留

	TransferValueInUsd float64 `json:"transferValueInUsd" bson:"transferValueInUsd"`

	BlockNo  uint64 `json:"blockNo" bson:"blockNo"`   // 区块号
	TxHash   string `json:"txHash" bson:"txHash"`     // 交易哈希
	Position uint   `json:"position" bson:"position"` // 交易在本区块中的序号

	LogIndexWithTx string `json:"-" bson:"logIndexWithTx"` // tx hash 以及 log 在本区块中的序号，以作为唯一标识

	TransferType int `json:"transferType" bson:"transferType"`

	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // transfer时间

	CreatedAt time.Time `json:"-" bson:"createdAt"`
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
		Keys:    bson.D{{Key: "amount", Value: 1}},
		Options: options.Index().SetName("amount_index"),
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
	{
		Keys:    bson.D{{Key: "transferType", Value: 1}},
		Options: options.Index().SetName("transferType_index"),
	},
}
