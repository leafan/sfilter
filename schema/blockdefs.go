package schema

import (
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Transaction struct {
	OriginTx *types.Transaction // 原始transaction信息

	Type int32 // 类型，如swap等，可以看成保留字段

	Receipt *types.Receipt
}

type Block struct {
	Block     *types.Block
	BlockNo   uint64
	BlockTime time.Time

	TxNums         int
	PairCreatedNum int
	VolumeByUsd    float64
	EthPrice       float64

	Transactions []*Transaction
}

// 该表的目的是确认是否已经被处理, 防止重复

type BlockProceeded struct {
	BlockNo   int64   `json:"blockNo" bson:"blockNo"`     // 区块号
	Hash      string  `json:"hash" bson:"hash"`           // 哈希
	BlockTime int64   `json:"blockTime" bson:"blockTime"` // 区块打包时间
	EthPrice  float64 `json:"ethPrice" bson:"ethPrice"`   // eth价格 by usd. 每个区块从usdc/eth v3 pool中获取

	TxNums      int     `json:"txNums" bson:"txNums"`
	VolumeByUsd float64 `json:"volumeByUsd" bson:"volumeByUsd"`

	Status    int       `json:"status" bson:"status"` // 状态，目前未使用
	CreatedAt time.Time `json:"-" bson:"createdAt"`   // 创建时间
}

var BlockProceededIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index"),
	},
	{
		Keys:    bson.D{{Key: "hash", Value: 1}},
		Options: options.Index().SetName("hash_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "blockNo", Value: 1}},
		Options: options.Index().SetName("blockNo_index"),
	},
}
