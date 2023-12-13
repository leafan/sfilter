package utils

import (
	"bytes"
	"encoding/gob"
	"math/big"
)

const INFINITE_CHANGE = 1

// 固定的quote币种列表
// wbtc去掉, 因为区块中只有eth价格
var QuoteEthCoinList = []string{
	"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // weth
	// "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", // wbtc
}

var QuoteUsdCoinList = []string{
	"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // usdc
	"0xdAC17F958D2ee523a2206206994597C13D831ec7", // usdt
	"0x6B175474E89094C44Da98b954EedeAC495271d0F", // dai
}

var CommonRuneAddresses = []string{
	"0x0000000000000000000000000000000000000000",
	"0x000000000000000000000000000000000000dEaD",
}

func CheckExistString(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

// 注意: src dst需要用指针
func DeepCopy(src, dst interface{}) error {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(src); err != nil {
		return err
	}

	return gob.NewDecoder(&buffer).Decode(dst)
}

// 计算百分比
func CalcChange(now, last float64) float32 {
	if last == 0 {
		return INFINITE_CHANGE
	}

	delta := now - last

	return float32(delta / now)
}

// 优先 usd; 优先 token1
func GetQuoteToken(token0, token1 string) string {
	var quoteToken string

	if CheckExistString(token1, QuoteUsdCoinList) {
		quoteToken = token1
	} else if CheckExistString(token0, QuoteUsdCoinList) {
		quoteToken = token0
	} else if CheckExistString(token1, QuoteEthCoinList) {
		quoteToken = token1
	} else if CheckExistString(token0, QuoteEthCoinList) {
		quoteToken = token0
	} else {
		quoteToken = token1
	}

	return quoteToken
}

// 根据一个token的amount, 计算出其usd值
func CalculateVolumeInUsd(token string, amount *big.Int, decimal uint8, ethPrice float64) float64 {
	var base float64

	if CheckExistString(token, QuoteUsdCoinList) {
		base = 1
	} else if CheckExistString(token, QuoteEthCoinList) {
		base = ethPrice
	} else {
		return 0
	}

	// 由于到这里计算的一定是价值币, 因此不会有奇葩币, 乘以 1e9 即可
	tokenExponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)-9), nil)
	amount = amount.Div(amount, tokenExponent)

	return float64(amount.Int64()) * base / 1e9
}

func IsValueCoin(token string) bool {
	if CheckExistString(token, QuoteUsdCoinList) {
		return true
	} else if CheckExistString(token, QuoteEthCoinList) {
		return true
	}

	return false
}

// 一些常见的销毁地址
func IsDeadAddress(token string) bool {
	if CheckExistString(token, CommonRuneAddresses) {
		return true
	}

	return false
}
