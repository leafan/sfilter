package schema

import (
	"log"

	"github.com/ethereum/go-ethereum/core/types"
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

// debug
func PrintOneBlock(block *Block) {
	log.Println("Number        : ", block.Block.Number())
	log.Println("Hash            : ", block.Block.Hash().Hex())

	for i, tx := range block.Transactions {
		log.Printf("tx %d hash: %v\n", i, tx.OriginTx.Hash())
		log.Printf("tx %d log count: %d \n", i, len(tx.Receipt.Logs))

		log.Printf("tx %d log[0]: %v\n", i, tx.Receipt.Logs[0])
	}
}
