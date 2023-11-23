package handler

import (
	"log"
	"sfilter/schema"
	"sfilter/services/pair"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandlePairLogic(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					handlePairCreated(block, _log, mongodb)
				}
			}
		}
	}
}

func handlePairCreated(block *schema.Block, _log *types.Log, mongodb *mongo.Client) {
	_pair := parsePairCreatedEvent(block, _log)
	if _pair != nil {
		log.Printf("[ handlePairCreated ] pair: %v, tx: %v\n\n", _pair, _log.TxHash)

		// 插入或更新pair
		pair.UpSertPairCreatedInfo(_pair, mongodb)
	}
}
