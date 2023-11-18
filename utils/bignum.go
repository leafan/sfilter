package utils

import "math/big"

func GetBigIntOrZero(number string) *big.Int {
	bigNum, ok := new(big.Int).SetString(number, 10)
	if !ok {
		bigNum = big.NewInt(0)
	}

	return bigNum
}
