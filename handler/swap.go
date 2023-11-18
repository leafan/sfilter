package handler

import (
	"fmt"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	service_swap "sfilter/services/swap"
	"sfilter/utils"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleSwap(block *schema.Block, mongodb *mongo.Client) int {
	swapNum := 0

	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					_type := checkSwapEvent(_log.Topics)

					// 发现有swap交易
					if _type > 0 {
						// log.Printf("[ Swap_handle ] swap tx now. type: %v, tx: %v\n\n", _type, tx.OriginTx.Hash())

						swap := newSwapStruct(block, _log, tx)
						if swap == nil {
							// new有错误, continue掉
							continue
						}
						swap.SwapType = _type

						if _type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
							updateUniV2Swap(swap, _log)
						} else if _type == schema.SWAP_EVENT_UNISWAPV3_LIKE {
							updateUniV3Swap(swap, _log)
						}

						updateExtraInfo(swap)

						handleOneSwap(swap, mongodb)

						swapNum++
					}
				}

			}
		}
	}

	return swapNum
}

// 本来都是协程进来, 这里不开协程了
func handleOneSwap(swap *schema.Swap, mongodb *mongo.Client) {
	service_swap.UpdateKline(swap, mongodb)

	service_swap.UpdateKOLTxTrends(swap, mongodb)

	service_swap.SaveSwapTx(swap, mongodb)
}

func updateExtraInfo(swap *schema.Swap) {
	// 找到quoteToken, 更新 VolumeInUsd.
	// 如果quoteToken为eth, 则乘以区块中eth价格; 如果为u, 直接加; 其他情况为0
	quoteToken := swap.Token1
	if swap.MainToken == swap.Token1 {
		quoteToken = swap.Token0
	}

	volume := utils.GetBigIntOrZero(swap.AmountOfMainToken)
	volumeInUsd := volume.Mul(volume, utils.GetBigIntOrZero(swap.Price))

	// price有乘以1e18, 要去掉
	volumeInUsd = volumeInUsd.Div(volumeInUsd, big.NewInt(1e18))

	if checkExistString(quoteToken, config.QuoteUsdCoinList) {
		swap.VolumeInUsd = volumeInUsd.String()
	} else if checkExistString(quoteToken, config.QuoteEthCoinList) {
		ethPrice := big.NewInt(int64(swap.CurrentEthPrice * 1e8)) // float转成int, 乘以1e8防止丢精度
		volumeInUsd = volumeInUsd.Mul(volumeInUsd, ethPrice)      // 乘以eth价格

		swap.VolumeInUsd = volumeInUsd.Div(volumeInUsd, big.NewInt(1e8)).String()
	} else {
		swap.VolumeInUsd = "0"
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

		SwapTime: int64(block.Block.Time()),
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
	} else {
		log.Printf("[ newSwapStruct ] wrong pair here. addr: %v, tx: %v\n", swap.PairAddr, swap.TxHash)
		return nil
	}

	swap.LogIndexWithTx = fmt.Sprintf("%s_%d", _log.TxHash.String(), _log.Index)
	swap.CurrentEthPrice = block.EthPrice

	return &swap
}
