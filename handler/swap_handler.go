package handler

import (
	"context"
	"log"
	"sfilter/schema"

	"go.mongodb.org/mongo-driver/mongo"
)

func Swap_handle(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					_type := schema.CheckSwapEvent(_log.Topics)

					// 发现有swap交易
					if _type > 0 {
						log.Printf("[ Swap_handle ] swap tx now. type: %v, tx: %v\n\n", _type, tx.OriginTx.Hash())

						var swap *schema.Swap

						if _type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
							swap = handleUniV2Swap(_log, tx)
						} else if _type == schema.SWAP_EVENT_UNISWAPV3_LIKE {
							swap = handleUniV3Swap(_log, tx)
						}

						if swap != nil {
							saveSwapInfo(swap, mongodb)
						}
					}
				}

			}
		}
	}
}

func saveSwapInfo(swap *schema.Swap, mongodb *mongo.Client) {
	collection := mongodb.Database("deepeye").Collection("swap")

	result, err := collection.InsertOne(context.Background(), swap)
	if err != nil {
		log.Printf("[ saveSwapInfo ] InsertOne error: %v\n", err)
	}

	log.Printf("Inserted document ID: %v\n", result.InsertedID)
}
