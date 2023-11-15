package handler

import (
	"fmt"
	"log"
	"math/big"
	"sfilter/schema"
	"sfilter/services/chain"
	service_swap "sfilter/services/swap"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleSwap(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					_type := checkSwapEvent(_log.Topics)

					// 发现有swap交易
					if _type > 0 {
						// log.Printf("[ Swap_handle ] swap tx now. type: %v, tx: %v\n\n", _type, tx.OriginTx.Hash())

						swap := newSwapStruct(block, _log, tx)
						swap.SwapType = _type

						if _type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
							updateUniV2Swap(swap, _log)
						} else if _type == schema.SWAP_EVENT_UNISWAPV3_LIKE {
							updateUniV3Swap(swap, _log)
						}

						handleOneSwap(swap, mongodb)
					}
				}

			}
		}
	}
}

func newSwapStruct(block *schema.Block, _log *types.Log, tx *schema.Transaction) *schema.Swap {
	swap := schema.Swap{
		BlockNo:  _log.BlockNumber,
		TxHash:   _log.TxHash.String(),
		Position: _log.TxIndex,

		PairAddr: _log.Address.String(),

		GasPrice: tx.Receipt.EffectiveGasPrice.String(),

		OperatorNonce: tx.OriginTx.Nonce(),

		SwapTime: time.Unix(int64(block.Block.Time()), 0),
	}

	effectiveGasPrice := big.NewInt(int64(tx.Receipt.GasUsed))
	effectiveGasPrice = effectiveGasPrice.Mul(effectiveGasPrice, tx.Receipt.EffectiveGasPrice)
	swap.GasInEth = effectiveGasPrice.String()

	// 解析发送者地址
	sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
	if err != nil {
		log.Printf("[ addBasicFields ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
	} else {
		swap.Operator = sender.String()
	}

	// 获取 token0, token1
	pair, err := chain.GetPairInfo(swap.PairAddr)
	if err == nil {
		swap.Token0 = pair.Token0
		swap.Token1 = pair.Token1
	}

	swap.LogIndexWithTx = fmt.Sprintf("%s_%d", _log.TxHash.String(), _log.Index)
	swap.CurrentEthPrice = block.EthPrice

	return &swap
}

func handleOneSwap(swap *schema.Swap, mongodb *mongo.Client) {
	go service_swap.UpdateKline(swap, mongodb)
	go service_swap.UpdateTxTrends(swap, mongodb)
	go service_swap.UpdateKOLTxTrends(swap, mongodb)

	go service_swap.SaveSwapTx(swap, mongodb)
}
