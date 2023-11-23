package chain

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

/*
refer: https://etherscan.io/tx/0x414eb36cfea5fb34e068a79196480e20caf021fb0921cf148f5aa0967faffe1d#eventlog

步骤:
1. 根据tokenId 调用 positions 获得 token0, token1, fee
2. 获得factory地址(不写死是为了防止有人fork v3代码导致混乱)
3. 调用factory的 getPool 函数获得pool地址
*/
func GetUniV3PoolAddr(postionManagerAddr string, tokenId *big.Int) (string, error) {
	var pool string
	token0, token1, fee, err := getUniV3Position(postionManagerAddr, tokenId)
	if err != nil {
		return pool, err
	}

	addr, err := getSingleProp(postionManagerAddr, "factory", getClient(), nil)
	if err != nil {
		return pool, err
	}
	factory := addr.(common.Address)

	poolAddr, err := getPoolAddr(factory, token0, token1, fee)
	if err != nil {
		return pool, err
	}

	pool = poolAddr.String()
	return pool, nil
}

func getPoolAddr(factoryAddr, token0, token1 common.Address, fee *big.Int) (common.Address, error) {
	var pool common.Address

	abi := getAbi()

	data, err := abi.Pack("getPool", token0, token1, fee)
	if err != nil {
		log.Printf("[ getPoolAddr ] Pack data error. factory: %v, err: %v\n", factoryAddr, err)
		return pool, err
	}
	msg := ethereum.CallMsg{
		From: common.Address{},
		To:   &factoryAddr,
		Data: data,
	}

	ret, err := getClient().CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Printf("[ getPoolAddr ] CallContract error. addr: %v, factory: %v\n", factoryAddr, err)
		return pool, err
	}

	intr, err := abi.Methods["getPool"].Outputs.UnpackValues(ret)
	if err != nil {
		log.Printf("[ getPoolAddr ] UnpackValues error. factory: %v, err: %v\n", factoryAddr, err)
		return pool, err
	}

	pool = intr[0].(common.Address)

	return pool, nil
}

func getUniV3Position(postionManagerAddr string, tokenId *big.Int) (common.Address, common.Address, *big.Int, error) {
	var token0, token1 common.Address
	var fee *big.Int

	abi := getAbi()
	contractAddr := common.HexToAddress(postionManagerAddr)

	data, err := abi.Pack("positions", tokenId)
	if err != nil {
		log.Printf("[ getUniV3Position ] Pack data error. tokenId: %v, err: %v\n", tokenId, err)
		return token0, token1, fee, err
	}
	msg := ethereum.CallMsg{
		From: common.Address{},
		To:   &contractAddr,
		Data: data,
	}

	ret, err := getClient().CallContract(context.Background(), msg, nil)
	if err != nil {
		// 回溯历史的时候, 由于 tokenId可能已被burn, 这里是可能报错的
		log.Printf("[ getUniV3Position ] CallContract error. addr: %v, tokenId: %v, err: %v\n", postionManagerAddr, tokenId, err)
		return token0, token1, fee, err
	}

	intr, err := abi.Methods["positions"].Outputs.UnpackValues(ret)
	if err != nil {
		log.Printf("[ getUniV3Position ] UnpackValues error. tokenId: %v, err: %v\n", tokenId, err)
		return token0, token1, fee, err
	}

	token0, _ = intr[2].(common.Address)
	token1 = intr[3].(common.Address)
	fee = intr[4].(*big.Int)

	return token0, token1, fee, nil
}

func TEST_POOL() {
	a, b, c, err := getUniV3Position("0xC36442b4a4522E871399CD717aBDD847Ab11FE88", big.NewInt(609781))
	fmt.Printf("[ TEST_POOL ] token0: %v, token1: %v, fee: %v, err: %v\n", a, b, c, err)

	addr, err := GetUniV3PoolAddr("0xC36442b4a4522E871399CD717aBDD847Ab11FE88", big.NewInt(609781))
	fmt.Printf("[ TEST_POOL ] pool addr: %v, err: %v\n", addr, err)
}
