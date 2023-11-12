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
	Status  int // reserved, no use now..

	CreatedAt time.Time
}

// const blockProceededSaveTime = 60 * 60 * 24 * 30 // 30d
const blockProceededSaveTime = 10

var BlockProceededIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdat", Value: -1}},
		Options: options.Index().SetName("createdat_index").SetExpireAfterSeconds(blockProceededSaveTime),
	},
	{
		Keys:    bson.D{{Key: "hash", Value: 1}},
		Options: options.Index().SetName("hash_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "blockno", Value: 1}},
		Options: options.Index().SetName("blockno_index"),
	},
}
