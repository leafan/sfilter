package swap

import (
	"fmt"
	"sfilter/schema"

	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func Swap_handler(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					_type := schema.CheckSwapEvent(_log.Topics)

					// 发现有swap交易
					if _type > 0 {
						// log.Printf("[ Swap_handle ] swap tx now. type: %v, tx: %v\n\n", _type, tx.OriginTx.Hash())

						var swap *schema.Swap

						if _type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
							swap = handleUniV2Swap(_log, tx)
						} else if _type == schema.SWAP_EVENT_UNISWAPV3_LIKE {
							swap = handleUniV3Swap(_log, tx)
						}

						if swap != nil {
							// 更多处理
							swap.LogIndexWithTx = fmt.Sprintf("%s_%d", _log.TxHash.String(), _log.Index)
							swap.CreatedAt = time.Now()
							handleOneSwap(swap, mongodb)
						}
					}
				}

			}
		}
	}
}

func handleOneSwap(swap *schema.Swap, mongodb *mongo.Client) {
	go UpdateKline(swap, mongodb)
	go UpdateTxTrends(swap, mongodb)
	go UpdateKOLTxTrends(swap, mongodb)

	go SaveSwapTx(swap, mongodb)
}
