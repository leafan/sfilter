package handler

import (
	"log"
	"math/big"
	"sfilter/schema"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/common/math"
)

// v2 addr: https://etherscan.io/address/0x42d52847be255eacee8c3f96b3b223c0b3cc0438
// v3 addr: https://etherscan.io/address/0xea05d862e4c5cd0d3e660e0fcb2045c8dd4d7912

func CheckExistString(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

func updateUniV2Swap(swap *schema.Swap, _log *types.Log) {
	// 解析event中的sender和recipient
	swap.Sender = common.HexToAddress(_log.Topics[1].Hex()).String()
	swap.Recipient = common.HexToAddress(_log.Topics[2].Hex()).String()

	// 获取event中的4个amount
	amount0In := new(big.Int).SetBytes(_log.Data[:32])
	amount1In := new(big.Int).SetBytes(_log.Data[32:64])
	amount0Out := new(big.Int).SetBytes(_log.Data[64:96])
	amount1Out := new(big.Int).SetBytes(_log.Data[96:128])

	log.Printf("\n\n[ updateUniV2Swap ] debug... tx: %v, amount0In: %v, amount1In: %v, amount0Out: %v, amount1Out: %v\n\n", _log.TxHash, amount0In, amount1In, amount0Out, amount1Out)

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

}

func updateUniV3Swap(swap *schema.Swap, l *types.Log) {
	swap.Sender = common.HexToAddress(l.Topics[1].Hex()).String()
	swap.Recipient = common.HexToAddress(l.Topics[2].Hex()).String()

	// 获取event中的4个amount
	amount0 := new(big.Int).SetBytes(l.Data[0:32])
	amount1 := new(big.Int).SetBytes(l.Data[32:64])
	sqrtPriceX96 := new(big.Int).SetBytes(l.Data[64:96])
	liquidity := new(big.Int).SetBytes(l.Data[96:128])
	tick := new(big.Int).SetBytes(l.Data[128:])

	// 使用ethereum官方库判断正负数
	amount0 = math.S256(amount0)
	amount1 = math.S256(amount1)
	tick = math.S256(tick)

	log.Printf("\n\n[ updateUniV3Swap ] debug... tx: %v amount0: %v, amount1In: %v, sqrtPriceX96: %v, liquidity: %v, tick: %v\n\n", l.TxHash, amount0, amount1, sqrtPriceX96, liquidity, tick)

}
