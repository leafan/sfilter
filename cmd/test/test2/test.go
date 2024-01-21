package main

import (
	"fmt"
	"math/big"

	"github.com/daoleno/uniswapv3-sdk/examples/quoter/uniswapv3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		panic(err)
	}
	quoterAddress := common.HexToAddress("0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6")
	quoterContract, err := uniswapv3.NewQuoter(quoterAddress, client)
	if err != nil {
		panic(err)
	}

	token0 := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	token1 := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	fee := big.NewInt(500)
	amountIn := big.NewInt(2600 * 1e6)
	sqrtPriceLimitX96 := big.NewInt(0)

	amountOut, err := quoterContract.QuoterCaller.QuoteExactInputSingle(&bind.CallOpts{}, token0, token1, fee, amountIn, sqrtPriceLimitX96)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("amountOut: ", amountOut)
}
