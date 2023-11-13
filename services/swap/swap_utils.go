package swap

import (
	"sfilter/schema"

	"github.com/ethereum/go-ethereum/core/types"
)

func CheckExistString(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func handleUniV2Swap(_log *types.Log, tx *schema.Transaction) *schema.Swap {
	swap := &schema.Swap{
		TxHash:        tx.OriginTx.Hash().Hex(),
		OperatorNonce: tx.OriginTx.Nonce(),
		// Operator: tx.OriginTx.From(),
	}

	swap.PairAddr = _log.Address.String()
	/*
		swap.Token0,swap.Token1 = getPairInfo(swap.PairAddr) //获取pair的token信息
		if CheckExistString(swap.Token0,QuoteCoinList) {
			swap.MainToken = 1
		} else {
			swap.MainToken = 0
		}
		amount0In.SetString(hex.EncodeToString(_log.Data()[:64]), 16)
		amount1In.SetString(hex.EncodeToString(_log.Data()[64:64*2]), 16)
		amount0Out.SetString(hex.EncodeToString(_log.Data()[64*2:64*3]), 16)
		amount1Out.SetString(hex.EncodeToString(_log.Data()[64*3:64*4]), 16)
		if amount0In == 0 && amount1Out == 0 & amount1In > 0 && amount0Out > 0 {
			token0decimal = getDecimals(swap.Token0)
			token1decimal = getDecimals(swap.Token1)
			if swap.MainToken == 0 {
				swap.Price = amount0Out / amount1In / token0decimal  * token1decimal
				swap.AmountOfMainToken = amount0Out
				swap.Direction = 0
			} else {
				swap.Price = amount1In / amount0Out / token1decimal * token0decimal
				swap.AmountOfMainToken = amount1In
				swap.Direction = 1
			}
		}
		if amount1In == 0 && amount0Out == 0 & amount0In > 0 && amount1Out > 0 {
			token0decimal = getDecimals(swap.Token0)
			token1decimal = getDecimals(swap.Token1)
			if swap.MainToken == 0 {
				swap.Price = amount0In / amount1Out / token0decimal  * token1decimal
				swap.AmountOfMainToken = amount0In
				swap.Direction = 1
			} else {
				swap.Price = amount1Out / amount0In / token1decimal * token0decimal
				swap.AmountOfMainToken = amount1Out
				swap.Direction = 0
			}
		}
	*/

	// ...

	// log.Printf("[ handleUniV2Log ] swap: %v\n", swap)

	return swap
}

func handleUniV3Swap(_log *types.Log, tx *schema.Transaction) *schema.Swap {
	swap := &schema.Swap{
		TxHash: tx.OriginTx.Hash().Hex(),
		// Operator: tx.OriginTx.From(),
	}

	// log.Printf("[ handleUniV3Log ] swap: %v\n", swap)

	return swap
}
