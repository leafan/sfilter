package handler

import (
	"fmt"
	"sfilter/schema"
	"sfilter/services/liquidity"
	"sfilter/services/pair"
	"sfilter/utils"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// 先执行pair creat的操作
// 在执行 handle liquidity 动作
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

		event.LogIndexWithTx = fmt.Sprintf("%s_%d", event.EventTxHash, l.Index)

		event.UpdatedAt = time.Now()
		event.CreatedAt = time.Now()

		_pair, err := pair.GetPairInfo(event.PoolAddress, mongodb)
		if err != nil || _pair == nil {
			utils.Warnf("[ handleAddLiquidity ] no pair?!! err: %v, tx: %v\n", err, event.EventTxHash)
			return
		}

		// 修正流动性amount value
		updateLiquidityEventValue(event, _pair, block)

		// 修正流动性池子大小
		UpdatePoolLiquidity(_pair, mongodb, block)

		// 判断如果是第一次添加流动性, 则update pair的firstAdd字段
		if event.Direction == schema.DIRECTION_BUY_OR_ADD {
			// 同时需要确认我们也检测到了 pairCreat 事件, 否则可能是老pair
			if _pair.PairCreatedBlockNo > 0 {
				if _pair.FirstAddPoolBlockNo <= 0 || _pair.FirstAddPoolBlockNo > event.EventBlockNo {
					info := &schema.InfoOnPools{
						FirstAddPoolBlockNo: event.EventBlockNo,
						FirstAddPoolTime:    event.EventTime,
						FirstAddTxHash:      event.EventTxHash,
						FirstAddGasPrice:    event.EventGasPrice,
					}

					pair.UpdatePoolInfo(_pair.Address, info, mongodb)
				}

			}

		}

		liquidity.SaveLiquidityEvent(event, mongodb)
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
