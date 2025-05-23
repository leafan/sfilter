package utils

import (
	"bytes"
	"encoding/gob"
	"math/big"
	"sfilter/config"
)

const INFINITE_CHANGE = 1

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

// 获取maintoken
func GetMainToken(token0, token1 string) string {
	quoteToken := GetQuoteToken(token0, token1)

	if quoteToken == token0 {
		return token1
	}

	return token0
}

// 优先 usd; 优先 token1
func GetQuoteToken(token0, token1 string) string {
	var quoteToken string

	if CheckExistString(token1, config.QuoteUsdCoinList) {
		quoteToken = token1
	} else if CheckExistString(token0, config.QuoteUsdCoinList) {
		quoteToken = token0
	} else if CheckExistString(token1, config.QuoteEthCoinList) {
		quoteToken = token1
	} else if CheckExistString(token0, config.QuoteEthCoinList) {
		quoteToken = token0
	} else {
		quoteToken = token1
	}

	return quoteToken
}

// 根据一个token的amount, 计算出其usd值
func CalculateVolumeInUsd(token string, amount *big.Float, decimal uint8, ethPrice float64) float64 {
	var base float64

	if CheckExistString(token, config.QuoteUsdCoinList) {
		base = 1
	} else if CheckExistString(token, config.QuoteEthCoinList) {
		base = ethPrice
	} else {
		return 0
	}

	// 由于到这里计算的一定是价值币, 因此不会有奇葩币, 乘以 1e6 即可
	// 这里也不能乘过大的数, 容易造成  amount 溢出
	tokenExponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
	amount = amount.Mul(amount, big.NewFloat(1e6))
	amount = amount.Quo(amount, new(big.Float).SetInt(tokenExponent))

	amountFloat, _ := amount.Float64()

	return amountFloat * base / 1e6
}

func IsValueCoin(token string) bool {
	if CheckExistString(token, config.QuoteUsdCoinList) {
		return true
	} else if CheckExistString(token, config.QuoteEthCoinList) {
		return true
	}

	return false
}

// 一些常见的销毁地址
func IsDeadAddress(token string) bool {
	return CheckExistString(token, config.BlackHoleAddresses)
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
