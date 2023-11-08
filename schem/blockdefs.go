package schem

import (
	"log"

	"github.com/ethereum/go-ethereum/core/types"
)

type OneTransaction struct {
	OriginTx *types.Transaction // 原始transaction信息

	Type int32 // 类型，如swap等，可以看成保留字段

	Receipt *types.Receipt
}

type OneBlock struct {
	Block        *types.Block
	Transactions []*OneTransaction
}

func PrintOneBlock(block *OneBlock) {
	log.Println("Number        : ", block.Block.Number)
	log.Println("Hash            : ", block.Block.Hash().Hex())

	for i, tx := range block.Transactions {
		log.Printf("tx %d hash: \n", i, tx.OriginTx.Hash())
		log.Printf("tx %d log count: %d \n", i, len(tx.Receipt.Logs))
	}
}
