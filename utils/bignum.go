package utils

import "math/big"

func GetBigIntOrZero(number string) *big.Int {
	bigNum, ok := new(big.Int).SetString(number, 10)
	if !ok {
		bigNum = big.NewInt(0)
	}

	return bigNum
}

func GetBigFloatOrZero(number string) *big.Float {
	bigNum, ok := new(big.Float).SetString(number)
	if !ok {
		bigNum = big.NewFloat(0)
	}

	return bigNum
}
