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
	Block        *types.Block
	EthPrice     float64
	Transactions []*Transaction
}

const (
	SWAP_EVENT_UNKNOWN int = iota

	SWAP_EVENT_UNISWAPV2_LIKE // uniswapv2 like
	SWAP_EVENT_UNISWAPV3_LIKE // uniswapv3 like

)

// 该表的目的是确认是否已经被处理, 防止重复
type BlockProceeded struct {
	BlockNo int64
	Hash    string

	BlockTime uint64  // 区块打包时间
	EthPrice  float64 // eth价格 by usd. 不需要非常精准，就根据少数几个交易对更新即可

	Status int // reserved, no use atm..

	CreatedAt time.Time
}

var BlockProceededIndexModel = []mongo.IndexModel{
	{
		Keys: bson.D{{Key: "createdat", Value: -1}},
		// Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(config.BlockProceededSaveTime),
		Options: options.Index().SetName("createdat_index"),
	},
	{
		Keys:    bson.D{{Key: "hash", Value: 1}},
		Options: options.Index().SetName("hash_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "blockno", Value: 1}},
		Options: options.Index().SetName("blockno_index").SetUnique(true),
	},
}
