package handler

import (
	"fmt"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/pair"
	service_swap "sfilter/services/swap"
	"sfilter/utils"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleSwapAndKline(block *schema.Block, transferMap schema.TxTokenTransfersMap, mongodb *mongo.Client) []*schema.Swap {
	var swaps []*schema.Swap

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
							updateUniV2Swap(swap, _log, mongodb)
						} else if _type == schema.SWAP_EVENT_UNISWAPV3_LIKE {
							updateUniV3Swap(swap, _log, mongodb)
						}

						updateVolumeinfo(swap)

						updateTrader(swap, transferMap)

						handleOneSwapAndKline(swap, mongodb)

						updateBlockInfo(block, swap)

						swaps = append(swaps, swap)
					}
				}

			}
		}
	}

	return swaps
}

func updateBlockInfo(blk *schema.Block, swap *schema.Swap) {
	blk.TxNums++
	blk.VolumeByUsd += swap.VolumeInUsd
}

/*
找出真正的受益方或者trader

if buy:

	if Token.transfer.To == operator: (遍历该tx中的transfer记录)
		trader = operator
	else // 如使用1inch挂单, 第三方吃单
		if Token.transfer.To 不为合约的地址:
	trader = To // 第一个就作为To吧, 可能不精确; 另外如果是合约购买的情况，也忽略了

else if sell:

	if Token.transfer.From == operator: (遍历该tx中的transfer记录)
		trader = operator
	else
		if Token.transfer.From 不为合约的地址:
			trader = Token.transfer.From // 如果是操作, 肯定是合约之间流转

如果上述没有找到，报错再分析，同时trader为null，表示作废
*/
func updateTrader(swap *schema.Swap, transferMap schema.TxTokenTransfersMap) {
	if swap.Direction != schema.DIRECTION_BUY_OR_ADD && swap.Direction != schema.DIRECTION_SELL_OR_DECREASE {
		return
	}

	key := fmt.Sprintf("%v_%v", swap.TxHash, swap.MainToken)

	transfers, ok := transferMap[key]
	if !ok {
		return
	}

	for _, transfer := range transfers {
		// log.Printf("[ updateTrader ] transfer: %v, key: %v\n", transfer, key)
		if swap.Direction == schema.DIRECTION_BUY_OR_ADD {
			if transfer.To == swap.Operator {
				swap.Trader = swap.Operator
				break
			}
		} else {
			if transfer.From == swap.Operator {
				swap.Trader = swap.Operator
				break
			}
		}
	}

	// 如果还没找到, 则遍历该token的transfer地址, 找到金额最大的不为contract的地址
	if swap.Trader == "" {
		biggestAmount := big.NewInt(0)

		// 为了防止通缩币向个人转账导致误判, 这里找出转账金额最大的那个
		for _, transfer := range transfers {
			address := transfer.To
			if swap.Direction == schema.DIRECTION_SELL_OR_DECREASE {
				address = transfer.From
			}

			if !chain.IsContract(address) && transfer.AmountBigInt.Cmp(biggestAmount) > 0 {
				swap.Trader = address
				biggestAmount = transfer.AmountBigInt // 一直更新until找到最大的
			}

		}
	}

}

// 本来都是协程进来, 这里不开协程了
func handleOneSwapAndKline(swap *schema.Swap, mongodb *mongo.Client) {
	// 应该先保存, 如果保存失败, 则说明不需要往后更新数据了
	err := service_swap.SaveSwapTx(swap, mongodb)

	// 数据为最近一周才update kline
	if err == nil && (time.Since(swap.SwapTime).Seconds() < config.SecondsForOneWeek) {
		UpdateKlines(swap, mongodb)
		service_swap.UpdateKOLTxTrends(swap, mongodb)
	}
}

func updateVolumeinfo(swap *schema.Swap) {
	// 找到quoteToken, 更新 VolumeInUsd.
	// 如果quoteToken为eth, 则乘以区块中eth价格; 如果为u, 直接加; 其他情况为0
	quoteToken := swap.Token1
	if swap.MainToken == swap.Token1 {
		quoteToken = swap.Token0
	}

	volume := utils.GetBigIntOrZero(swap.AmountOfMainToken)
	volumeInUsd := volume.Mul(volume, utils.GetBigIntOrZero(swap.Price))

	// amount 乘了1e36 的基数
	volumeInUsd = volumeInUsd.Div(volumeInUsd, config.AmountBaseFactor1e36)

	// price有乘以1e18, 要去掉, 此时由于已经是eth或者usd, 所以除以1e9防止误差即可
	volumeInUsd = volumeInUsd.Div(volumeInUsd, big.NewInt(1e9))

	floatWithoutDecimal, _ := new(big.Float).SetInt(volumeInUsd).Float64()

	if utils.CheckExistString(quoteToken, config.QuoteUsdCoinList) {
		swap.VolumeInUsd = floatWithoutDecimal / 1e9
	} else if utils.CheckExistString(quoteToken, config.QuoteEthCoinList) {
		swap.VolumeInUsd = floatWithoutDecimal * swap.CurrentEthPrice / 1e9
	} else {
		swap.VolumeInUsd = 0
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
		log.Printf("[ newSwapStruct ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
	} else {
		swap.Operator = sender.String()
	}

	// 获取 token0, token1
	pair, err := pair.GetPairInfoForRead(swap.PairAddr)
	if err == nil {
		swap.Token0 = pair.Token0
		swap.Token1 = pair.Token1
		swap.PairName = pair.PairName
	} else {
		log.Printf("[ newSwapStruct ] wrong pair here. addr: %v, tx: %v\n", swap.PairAddr, swap.TxHash)
		return nil
	}

	swap.LogIndexWithTx = fmt.Sprintf("%s_%d", _log.TxHash.String(), _log.Index)
	swap.CurrentEthPrice = block.EthPrice

	return &swap
}
