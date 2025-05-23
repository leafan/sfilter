package handler

import (
	//"fmt"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"strings"

	"sfilter/services/chain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/ethereum/go-ethereum/common/math"
)

// v2 addr: https://etherscan.io/address/0x42d52847be255eacee8c3f96b3b223c0b3cc0438
// v3 addr: https://etherscan.io/address/0xea05d862e4c5cd0d3e660e0fcb2045c8dd4d7912

func updateUniV2Swap(swap *schema.Swap, _log *types.Log, mongodb *mongo.Client) {
	// 解析event中的sender和recipient
	swap.Sender = common.HexToAddress(_log.Topics[1].Hex()).String()
	swap.Recipient = common.HexToAddress(_log.Topics[2].Hex()).String()

	// 获取event中的4个amount
	amount0In := new(big.Int).SetBytes(_log.Data[:32])
	amount1In := new(big.Int).SetBytes(_log.Data[32:64])
	amount0Out := new(big.Int).SetBytes(_log.Data[64:96])
	amount1Out := new(big.Int).SetBytes(_log.Data[96:128])

	// log.Printf("\n\n[ updateUniV2Swap ] debug... tx: %v, amount0In: %v, amount1In: %v, amount0Out: %v, amount1Out: %v\n\n", _log.TxHash, amount0In, amount1In, amount0Out, amount1Out)

	// 取出token0和token1的decimals
	token0, err0 := chain.GetTokenInfo(swap.Token0, mongodb)
	token1, err1 := chain.GetTokenInfo(swap.Token1, mongodb)
	if err0 != nil || err1 != nil {
		log.Printf("[ updateUniV2Swap ] GetTokenInfo error! err0: %v, err1: %v, tx: %v\n", err0, err1, _log.TxHash)
		return
	}

	quoteToken := utils.GetQuoteToken(swap.Token0, swap.Token1)
	if swap.Token0 == quoteToken {
		swap.MainToken = swap.Token1
	} else {
		swap.MainToken = swap.Token0
	}

	swap.Decimal0 = token0.Decimal
	swap.Decimal1 = token1.Decimal

	// AmountOfMainToken 要除以其decimal, 相当于实际上的amount
	// 为了考虑精度, 乘以1e36(有些屌丝token的decimal为30+..)

	if (amount0In.Cmp(big.NewInt(0)) == 0 || amount1Out.Cmp(big.NewInt(0)) == 0) && amount1In.Cmp(big.NewInt(0)) > 0 && amount0Out.Cmp(big.NewInt(0)) > 0 {
		token0Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token0.Decimal)), nil)
		token1Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token1.Decimal)), nil)
		if swap.MainToken == swap.Token0 {
			swap.AmountOfMainBig = amount0Out.String()
			swap.AmountOfQuoteBig = amount1In.String()

			calculatePrice := new(big.Int).Mul(amount1In, token0Exponent)
			calculatePrice.Mul(calculatePrice, config.BaseFactor1e18)
			calculatePrice.Div(calculatePrice, token1Exponent)
			calculatePrice.Div(calculatePrice, amount0Out)

			tmpFloat, _ := calculatePrice.Float64()
			swap.Price = tmpFloat / 1e18

			amount0Out = amount0Out.Mul(amount0Out, config.BaseFactor1e18)
			amount0Out = amount0Out.Div(amount0Out, token0Exponent)

			tmpFloat, _ = amount0Out.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_BUY_OR_ADD

			//log.Println("[ updateUniV2Swap ] debug... price : ", amount0Out, amount1In, token0Exponent, token1Exponent, calculatePrice)
		} else if amount1In.Cmp(big.NewInt(0)) > 0 {
			swap.AmountOfMainBig = amount1In.String()
			swap.AmountOfQuoteBig = amount0Out.String()

			calculatePrice := new(big.Int).Mul(amount0Out, token1Exponent)
			calculatePrice.Mul(calculatePrice, (config.BaseFactor1e18))
			calculatePrice.Div(calculatePrice, token0Exponent)
			calculatePrice.Div(calculatePrice, amount1In)

			tmpFloat, _ := calculatePrice.Float64()
			swap.Price = tmpFloat / 1e18

			amount1In = amount1In.Mul(amount1In, config.BaseFactor1e18)
			amount1In = amount1In.Div(amount1In, token1Exponent)

			tmpFloat, _ = amount1In.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_SELL_OR_DECREASE

			//log.Println("[ updateUniV2Swap ] debug... price :\n\n", amount1In, amount0Out, token0Exponent, token1Exponent, calculatePrice)
		}
	}

	if (amount1In.Cmp(big.NewInt(0)) == 0 || amount0Out.Cmp(big.NewInt(0)) == 0) && amount0In.Cmp(big.NewInt(0)) > 0 && amount1Out.Cmp(big.NewInt(0)) > 0 {
		token0Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token0.Decimal)), nil)
		token1Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token1.Decimal)), nil)
		if swap.MainToken == swap.Token0 {
			swap.AmountOfMainBig = amount0In.String()
			swap.AmountOfQuoteBig = amount1Out.String()

			calculatePrice := new(big.Int).Mul(amount1Out, token0Exponent)
			calculatePrice.Mul(calculatePrice, config.BaseFactor1e18)
			calculatePrice.Div(calculatePrice, token1Exponent)
			calculatePrice.Div(calculatePrice, amount0In)

			tmpFloat, _ := calculatePrice.Float64()
			swap.Price = tmpFloat / 1e18

			amount0In = amount0In.Mul(amount0In, config.BaseFactor1e18)
			amount0In = amount0In.Div(amount0In, token0Exponent)

			tmpFloat, _ = amount0In.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_SELL_OR_DECREASE

			// log.Println("[ updateUniV2Swap ] debug... price :\n\n", amount0Out, amount1In, token0Exponent, token1Exponent, calculatePrice)
		} else if amount1Out.Cmp(big.NewInt(0)) > 0 {
			swap.AmountOfMainBig = amount1Out.String()
			swap.AmountOfQuoteBig = amount0In.String()

			calculatePrice := new(big.Int).Mul(amount0In, token1Exponent)
			calculatePrice.Mul(calculatePrice, config.BaseFactor1e18)
			calculatePrice.Div(calculatePrice, token0Exponent)
			calculatePrice.Div(calculatePrice, amount1Out)

			tmpFloat, _ := calculatePrice.Float64()
			swap.Price = tmpFloat / 1e18

			amount1Out = amount1Out.Mul(amount1Out, config.BaseFactor1e18)
			amount1Out = amount1Out.Div(amount1Out, token1Exponent)

			tmpFloat, _ = amount1Out.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_BUY_OR_ADD

			// log.Println("[ updateUniV2Swap ] debug... price :\n\n", amount1In, amount0Out, token0Exponent, token1Exponent, calculatePrice)
		}
	}

	// log.Printf("[ handleUniV2Log ] swap: %v\n", swap)

}

func updateUniV3Swap(swap *schema.Swap, l *types.Log, mongodb *mongo.Client) {
	swap.Sender = common.HexToAddress(l.Topics[1].Hex()).String()
	swap.Recipient = common.HexToAddress(l.Topics[2].Hex()).String()

	// 获取event中的data
	amount0 := new(big.Int).SetBytes(l.Data[0:32])
	amount1 := new(big.Int).SetBytes(l.Data[32:64])
	// sqrtPriceX96 := new(big.Int).SetBytes(l.Data[64:96])
	// liquidity := new(big.Int).SetBytes(l.Data[96:128])
	// tick := new(big.Int).SetBytes(l.Data[128:])

	// 使用ethereum官方库判断正负数
	amount0 = math.S256(amount0)
	amount1 = math.S256(amount1)
	// log.Println("\n\n[ updateUniV3Swap ] debug... ", amount0, amount1)
	// tick = math.S256(tick)

	// log.Printf("\n\n[ updateUniV3Swap ] debug... tx: %v amount0: %v, amount1In: %v, sqrtPriceX96: %v, liquidity: %v, tick: %v\n\n", l.TxHash, amount0, amount1, sqrtPriceX96, liquidity, tick)

	// 取出token0和token1的decimals
	token0, err0 := chain.GetTokenInfo(swap.Token0, mongodb)
	token1, err1 := chain.GetTokenInfo(swap.Token1, mongodb)
	if err0 != nil || err1 != nil {
		log.Printf("[ updateUniV3Swap ] GetTokenInfo error! err0: %v, err1: %v\n", err0, err1)
		return
	}

	quoteToken := utils.GetQuoteToken(swap.Token0, swap.Token1)
	if swap.Token0 == quoteToken {
		swap.MainToken = swap.Token1
	} else {
		swap.MainToken = swap.Token0
	}

	swap.Decimal0 = token0.Decimal
	swap.Decimal1 = token1.Decimal

	token0Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token0.Decimal)), nil)
	token1Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token1.Decimal)), nil)
	if swap.MainToken == swap.Token0 {
		swap.AmountOfMainBig = amount0.String()
		swap.AmountOfQuoteBig = amount1.String()
		if amount0.Cmp(big.NewInt(0)) > 0 && amount1.Cmp(big.NewInt(0)) < 0 {
			amount1 = new(big.Int).Sub(big.NewInt(0), amount1)
			calculatePrice := new(big.Int).Mul(amount1, token0Exponent)
			calculatePrice.Mul(calculatePrice, config.BaseFactor1e18)
			calculatePrice.Div(calculatePrice, token1Exponent)
			calculatePrice.Div(calculatePrice, amount0)

			tmpFloat, _ := calculatePrice.Float64()

			swap.Price = tmpFloat / 1e18
			amount0 = amount0.Mul(amount0, config.BaseFactor1e18)
			amount0 = amount0.Div(amount0, token0Exponent)

			tmpFloat, _ = amount0.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_SELL_OR_DECREASE
		} else if amount0.Cmp(big.NewInt(0)) < 0 && amount1.Cmp(big.NewInt(0)) > 0 {

			amount0 = new(big.Int).Sub(big.NewInt(0), amount0)
			calculatePrice := new(big.Int).Mul(amount1, token0Exponent)
			calculatePrice.Mul(calculatePrice, config.BaseFactor1e18)
			calculatePrice.Div(calculatePrice, token1Exponent)
			calculatePrice.Div(calculatePrice, amount0)

			tmpFloat, _ := calculatePrice.Float64()
			swap.Price = tmpFloat / 1e18

			amount0 = amount0.Mul(amount0, config.BaseFactor1e18)
			amount0 = amount0.Div(amount0, token0Exponent)

			tmpFloat, _ = amount0.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_BUY_OR_ADD
		}
	}

	if swap.MainToken == swap.Token1 {
		swap.AmountOfMainBig = amount1.String()
		swap.AmountOfQuoteBig = amount0.String()
		if amount0.Cmp(big.NewInt(0)) > 0 && amount1.Cmp(big.NewInt(0)) < 0 {

			amount1 = new(big.Int).Sub(big.NewInt(0), amount1)
			calculatePrice := new(big.Int).Mul(amount0, token1Exponent)
			calculatePrice.Mul(calculatePrice, config.BaseFactor1e18)
			calculatePrice.Div(calculatePrice, token0Exponent)
			calculatePrice.Div(calculatePrice, amount1)

			tmpFloat, _ := calculatePrice.Float64()
			swap.Price = tmpFloat / 1e18

			amount1 = amount1.Mul(amount1, config.BaseFactor1e18)
			amount1 = amount1.Div(amount1, token1Exponent)

			tmpFloat, _ = amount1.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_BUY_OR_ADD
		} else if amount0.Cmp(big.NewInt(0)) < 0 && amount1.Cmp(big.NewInt(0)) > 0 {

			amount0 = new(big.Int).Sub(big.NewInt(0), amount0)
			calculatePrice := new(big.Int).Mul(amount0, token1Exponent)
			calculatePrice.Mul(calculatePrice, config.BaseFactor1e18)
			calculatePrice.Div(calculatePrice, token0Exponent)
			calculatePrice.Div(calculatePrice, amount1)

			tmpFloat, _ := calculatePrice.Float64()
			swap.Price = tmpFloat / 1e18

			amount1 = amount1.Mul(amount1, config.BaseFactor1e18)
			amount1 = amount1.Div(amount1, token1Exponent)

			tmpFloat, _ = amount1.Float64()
			swap.AmountOfMainToken = tmpFloat / 1e18

			swap.Direction = schema.DIRECTION_SELL_OR_DECREASE
		}
	}

	// log.Println("[ updateUniV3Swap ] update success!", swap.Price, swap.AmountOfMainToken, swap.Direction, l.TxHash)

}

func checkSwapEvent(topics []common.Hash) int {
	if isUniswapSwapV2Event(topics) {
		return schema.SWAP_EVENT_UNISWAPV2_LIKE
	}

	if isUniswapSwapV3Event(topics) {
		return schema.SWAP_EVENT_UNISWAPV3_LIKE
	}

	return schema.SWAP_EVENT_UNKNOWN
}

func isUniswapSwapV2Event(topics []common.Hash) bool {
	if len(topics) == 3 {
		return strings.EqualFold(topics[0].String(), "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822")
	}

	return false
}

func isUniswapSwapV3Event(topics []common.Hash) bool {
	if len(topics) == 3 {
		return strings.EqualFold(topics[0].String(), "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67")
	}

	return false
}
