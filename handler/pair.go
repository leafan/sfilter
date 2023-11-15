package handler

import (
	"log"
	"sfilter/schema"
	"sfilter/services/pair"

	"go.mongodb.org/mongo-driver/mongo"
)

func HandlePairCreated(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					pair := parsePairCreatedEvent(block, _log)

					if pair != nil {
						log.Printf("[ HandlePairCreated ] pair: %v, tx: %v\n\n", pair, tx.OriginTx.Hash())

						handlePairCreated(pair, mongodb)
					}
				}

			}
		}
	}
}

func handlePairCreated(_pair *schema.Pair, mongodb *mongo.Client) {
	// 插入或更新pair
	pair.UpSertPairInfo(_pair, mongodb)
}
