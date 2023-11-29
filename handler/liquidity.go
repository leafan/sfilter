package handler

import (
	"sfilter/schema"
	"sfilter/services/liquidity"
	"sfilter/services/pair"
	"sfilter/utils"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// 先执行pair creat的操作
func HandleLiquidityLogic(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					handleAddLiquidity(block, tx, _log, mongodb)
				}
			}
		}
	}
}

func handleAddLiquidity(block *schema.Block, tx *schema.Transaction, l *types.Log, mongodb *mongo.Client) {
	event := parseLiquidityEvent(tx, l)
	if event != nil {
		event.EventBlockNo = l.BlockNumber
		event.EventTime = time.Unix(int64(block.Block.Time()), 0)
		event.EventTxHash = l.TxHash.String()
		event.EventGasPrice = tx.Receipt.EffectiveGasPrice.String()

		event.UpdatedAt = time.Now()
		event.CreatedAt = time.Now()

		liquidity.SaveLiquidityEvent(event, mongodb)

		// 判断如果是第一次添加流动性, 则update pair的firstAdd字段
		if event.Direction == schema.DIRECTION_BUY_OR_ADD {
			_pair, err := pair.GetPairInfo(event.PoolAddress, mongodb)
			if err != nil {
				return
			}

			// 同时需要确认我们也检测到了 pairCreat 事件, 否则可能是老pair
			if _pair.PairCreatedBlockNo > 0 {
				var update = false

				if _pair.FirstAddPoolBlockNo <= 0 || _pair.FirstAddPoolBlockNo > event.EventBlockNo {
					update = true
				}

				if _pair.FirstAddPoolBlockNo == event.EventBlockNo {
					oldGas := utils.GetBigIntOrZero(_pair.FirstAddGasPrice)
					newGas := utils.GetBigIntOrZero(event.EventGasPrice)

					if oldGas.Cmp(newGas) < 0 {
						update = true
					}
				}

				if update {
					info := &schema.InfoOnPools{
						FirstAddPoolBlockNo: event.EventBlockNo,
						FirstAddPoolTime:    event.EventTime,
						FirstAddTxHash:      event.EventTxHash,
						FirstAddGasPrice:    event.EventGasPrice,
					}

					// update
					pair.UpdatePoolInfo(_pair.Address, info, mongodb)
				}

			}

		}
	}
}

func parseLiquidityEvent(tx *schema.Transaction, l *types.Log) *schema.LiquidityEvent {
	var event *schema.LiquidityEvent

	event = parseUniV2AddLiquidity(l, tx)
	if event != nil {
		return event
	}

	event = parseUniV3AddLiquidity(l, tx)
	if event != nil {
		return event
	}

	event = parseUniV2RemoveLiquidity(l, tx)
	if event != nil {
		return event
	}

	event = parseUniV3RemoveLiquidity(l, tx)
	if event != nil {
		return event
	}

	return nil
}
