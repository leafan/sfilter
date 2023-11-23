package handler

import (
	"log"
	"math/big"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/liquidity"
	"sfilter/services/pair"
	"strings"
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
	event := parseLiquidityEvent(block, tx, l)
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

			if err == nil {
				// 同时需要确认我们也检测到了 pairCreat 事件, 否则可能是老pair
				if _pair.PairCreatedBlockNo > 0 && _pair.FirstAddPoolBlockNo <= 0 {
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

func parseLiquidityEvent(block *schema.Block, tx *schema.Transaction, l *types.Log) *schema.LiquidityEvent {
	var event *schema.LiquidityEvent

	// v2
	if len(l.Topics) == 2 && strings.EqualFold(l.Topics[0].String(), "0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f") {
		event = &schema.LiquidityEvent{
			PoolAddress: l.Address.String(),
			Direction:   schema.DIRECTION_BUY_OR_ADD, // add
			Type:        schema.SWAP_EVENT_UNISWAPV2_LIKE,
		}

		sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
		if err != nil {
			log.Printf("[ parseLiquidityEvent ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
		} else {
			event.Operator = sender.String()
		}

		event.Amount0 = new(big.Int).SetBytes(l.Data[0:32]).String()
		event.Amount1 = new(big.Int).SetBytes(l.Data[32:64]).String()
	}

	// uni v3 的 IncreaseLiquidity 函数
	// refer: https://etherscan.io/tx/0x414eb36cfea5fb34e068a79196480e20caf021fb0921cf148f5aa0967faffe1d#eventlog
	if len(l.Topics) == 2 && strings.EqualFold(l.Topics[0].String(), "0x3067048beee31b25b2f1681f88dac838c8bba36af25bfb2b7cf7473a5847e35f") {
		event = &schema.LiquidityEvent{
			Direction: schema.DIRECTION_BUY_OR_ADD, // add
			Type:      schema.SWAP_EVENT_UNISWAPV3_LIKE,
		}

		// v3 获取pair地址, 及其坑爹.. 这个破玩意儿
		var err error
		tokenId, ok := new(big.Int).SetString(l.Topics[1].Hex(), 0)
		if !ok {
			return nil
		}

		event.PoolAddress, err = chain.GetUniV3PoolAddr(l.Address.String(), tokenId)
		if err != nil {
			// log.Printf("[ parseLiquidityEvent ] GetUniV3PoolAddr  err: %v, tx: %v\n", err, tx.OriginTx.Hash())
			return nil
		}

		sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
		if err != nil {
			log.Printf("[ parseLiquidityEvent ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
		} else {
			event.Operator = sender.String()
		}

		event.Amount0 = new(big.Int).SetBytes(l.Data[32:64]).String()
		event.Amount1 = new(big.Int).SetBytes(l.Data[64:96]).String()
	}

	// uni v3的decraese
	// todo, 还需要把burn弄进来, 先不处理了

	return event
}
