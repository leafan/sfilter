package chain

import (
	"context"
	"fmt"
	"math/big"
	"sfilter/schema"
	"sfilter/utils"

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
		utils.Warnf("[ getPoolAddr ] Pack data error. factory: %v, err: %v", factoryAddr, err)
		return pool, err
	}
	msg := ethereum.CallMsg{
		From: common.Address{},
		To:   &factoryAddr,
		Data: data,
	}

	ret, err := getClient().CallContract(context.Background(), msg, nil)
	if err != nil {
		utils.Debugf("[ getPoolAddr ] CallContract error. addr: %v, factory: %v\n", factoryAddr, err)
		return pool, err
	}

	intr, err := abi.Methods["getPool"].Outputs.UnpackValues(ret)
	if err != nil {
		utils.Warnf("[ getPoolAddr ] UnpackValues error. factory: %v, err: %v\n", factoryAddr, err)
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
		utils.Warnf("[ getUniV3Position ] Pack data error. tokenId: %v, err: %v\n", tokenId, err)
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
		utils.Debugf("[ getUniV3Position ] CallContract error. addr: %v, tokenId: %v, err: %v\n", postionManagerAddr, tokenId, err)
		return token0, token1, fee, err
	}

	intr, err := abi.Methods["positions"].Outputs.UnpackValues(ret)
	if err != nil {
		utils.Warnf("[ getUniV3Position ] UnpackValues error. tokenId: %v, err: %v\n", tokenId, err)
		return token0, token1, fee, err
	}

	token0, _ = intr[2].(common.Address)
	token1 = intr[3].(common.Address)
	fee = intr[4].(*big.Int)

	return token0, token1, fee, nil
}

func GetUniPoolType(poolAddr string) (int, error) {
	// uni v2 check
	_, err := getSingleProp(poolAddr, "price0CumulativeLast", getClient(), nil)
	if err == nil {
		return schema.SWAP_EVENT_UNISWAPV2_LIKE, nil
	}

	// uni v3 check
	_, err = getSingleProp(poolAddr, "maxLiquidityPerTick", getClient(), nil)
	if err == nil {
		return schema.SWAP_EVENT_UNISWAPV3_LIKE, nil
	}

	return 0, err
}

func TEST_POOL() {
	v2, err2 := GetUniPoolType("0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc")

	v3, err3 := GetUniPoolType("0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640")

	fmt.Printf("[ TEST_POOL ] v2: %v, v3: %v, err2: %v, err3: %v\n", v2, v3, err2, err3)
}
