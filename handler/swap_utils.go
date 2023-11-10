package handler

import (
	"log"
	"sfilter/schema"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

func handleUniV2Swap(_log *types.Log, tx *schema.Transaction) *schema.Swap {
	swap := &schema.Swap{
		Tx: tx.OriginTx.Hash().Hex(),
		// Operator: tx.OriginTx.From(),
		Receiver:  tx.OriginTx.To().String(),
		CreatedAt: time.Now(),
	}
	log.Printf("[ handleUniV2Log ] swap: %v\n", swap)

	return swap
}

func handleUniV3Swap(_log *types.Log, tx *schema.Transaction) *schema.Swap {
	swap := &schema.Swap{
		Tx: tx.OriginTx.Hash().Hex(),
		// Operator: tx.OriginTx.From(),
		Receiver:  tx.OriginTx.To().String(),
		CreatedAt: time.Now(),
	}

	log.Printf("[ handleUniV3Log ] swap: %v\n", swap)

	return swap
}
