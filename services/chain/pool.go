package chain

import (
	"context"
	"errors"
	"log"
	"math/big"

	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/daoleno/uniswapv3-sdk/examples/quoter/uniswapv3"
)

var chain_quoter *uniswapv3.Quoter

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

// UniV2 getReserves
func GetUniV2PairReserves(pair string) (*big.Int, *big.Int, error) {
	reserves, err := GetNoParamProp(pair, "getReserves")
	if err != nil {
		log.Printf("[ GetUniV2PairReserves ] getReserves error. pair: %v, err: %v\n", pair, err)
		return nil, nil, err
	}

	reserveArr, ok := reserves.([]interface{})
	if !ok {
		return nil, nil, errors.New("failed to convert reserves to []interface{}")
	}

	if len(reserveArr) < 2 {
		return nil, nil, errors.New("insufficient reserve values")
	}

	reserve0, ok := reserveArr[0].(*big.Int)
	if !ok {
		return nil, nil, errors.New("failed to convert reserve0 to *big.Int")
	}

	reserve1, ok := reserveArr[1].(*big.Int)
	if !ok {
		return nil, nil, errors.New("failed to convert reserve1 to *big.Int")
	}

	return reserve0, reserve1, nil
}

// uni v2
func GetUniV2SwapAmountOut(pair string, token0 string, tokenIn string, amountIn *big.Int) (*big.Int, error) {
	r0, r1, err := GetUniV2PairReserves(pair)
	if err != nil {
		return nil, err
	}

	var amountOut *big.Int

	if token0 == tokenIn {
		amountOut = getAmountOut(amountIn, r0, r1)
	} else {
		amountOut = getAmountOut(amountIn, r1, r0)
	}

	return amountOut, nil
}

func getAmountOut(amountIn *big.Int, reserveIn *big.Int, reserveOut *big.Int) *big.Int {
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	denominator := new(big.Int).Mul(reserveIn, big.NewInt(1000))
	denominator.Add(denominator, amountInWithFee)
	if denominator.Sign() == 0 {
		return new(big.Int)
	}
	return numerator.Div(numerator, denominator)
}

func GetUniV3SwapAmountOut(tokenIn, tokenOut string, fee, amountIn *big.Int) (*big.Int, error) {
	var err error
	if chain_quoter == nil {
		chain_quoter, err = uniswapv3.NewQuoter(common.HexToAddress(config.Quoter_Contract_Address), getClient())
		if err != nil {
			utils.Errorf("[ GetUniV3SwapAmountOut ] NewQuoter error: %v", err)
			return nil, err
		}
	}

	tokenInAddr := common.HexToAddress(tokenIn)
	tokenOutAddr := common.HexToAddress(tokenOut)

	amountOut, err := chain_quoter.QuoterCaller.QuoteExactInputSingle(&bind.CallOpts{}, tokenInAddr, tokenOutAddr, fee, amountIn, big.NewInt(0))
	if err != nil {
		utils.Warnf("[ GetUniV3SwapAmountOut ] call QuoteExactInputSingle error: %v", err)
		return nil, err
	}

	return amountOut, nil
}

func GetUniV2PoolTokenHackType(_pair *schema.Pair) int {
	// 如下假设我们合约有钱, 没钱也会误判..!

	// 先检测是否是正常币
	ret := getHackStatusFromContract(_pair, 1)
	if ret == 1 { // 如果执行失败, 说明是通缩币或者坑人币
		retDeflat := getHackStatusFromContract(_pair, 2)
		if retDeflat == 0 {
			// 此时一定为通缩币
			return schema.PAIR_MAINTOKEN_HACK_TYPE_DEFLAT
		} else if retDeflat == 1 {
			// 说明是坑人币
			return schema.PAIR_MAINTOKEN_HACK_TYPE_SCAM
		}
	} else if ret == 0 {
		// 正常币
		return schema.PAIR_MAINTOKEN_HACK_TYPE_NORMAL
	} else if ret == 3 {
		return schema.PAIR_MAINTOKEN_HACK_TYPE_EMPTY_BALANCE
	}

	// 无法检测, unknown
	return schema.PAIR_MAINTOKEN_HACK_TYPE_UNKNOWN
}

// 获取pair中的maintoken是否为通缩币、坑人币等
//
//	0表示检测合约通过; 1表示检测合约不通过; 2表示无法检测; 3 表示pair没钱
func getHackStatusFromContract(_pair *schema.Pair, divFactorInt int) int {
	hackCheckAddr := config.Hacker_Check_Contract_Address

	abi := getAbi()
	contractAddr := common.HexToAddress(hackCheckAddr)

	// 必须有一个是eth token, 否则认为是不识别类型
	if !utils.Contains(config.QuoteEthCoinList, _pair.Token0) && !utils.Contains(config.QuoteEthCoinList, _pair.Token1) {
		return 2
	}

	// 找出weth
	var tokenFrom, tokenTo string
	if utils.Contains(config.QuoteEthCoinList, _pair.Token0) {
		tokenFrom = _pair.Token0
		tokenTo = _pair.Token1
	} else {
		tokenFrom = _pair.Token1
		tokenTo = _pair.Token0
	}

	//先判断里面余额是否 >0, 否则合约肯定失败
	balanceOfMainToken, err := BalanceOf(_pair.Address, tokenTo)
	if err != nil || balanceOfMainToken.String() == "0" {
		return 3 // 余额不足
	}

	amount := big.NewInt(1e15)
	divFactor := big.NewInt(int64(divFactorInt))
	data, err := abi.Pack("hackTestForUniV2", common.HexToAddress(_pair.Address), common.HexToAddress(tokenFrom), common.HexToAddress(tokenTo), common.HexToAddress(_pair.Token0), amount, divFactor)
	if err != nil {
		utils.Warnf("[ getHackStatusFromContract ] Pack data error. address: %v, err: %v\n", _pair.Address, err)
		return 2
	}
	msg := ethereum.CallMsg{
		From: common.HexToAddress("0x821B711A51e1AfCaa68F02dD38B72E6e71d91A33"), // executor
		To:   &contractAddr,
		Data: data,
	}

	_, err = getClient().CallContract(context.Background(), msg, nil)
	if err != nil {
		// utils.Warnf("[ getHackStatusFromContract ] CallContract error. addr: %v, err: %v\n", _pair.Address, err)

		// 跑到这里说明是通缩币或者坑人币

		return 1
	}

	// 如果正常结束, 说明合约测试通过
	return 0
}

// 获取bsc链上的wbnb等价格
// 这是基础函数, 写挫一点, 不引入任何第三方接口
func GetBnbPriceByBsc(client *ethclient.Client) (float64, error) {
	poolAddr := "0x16b9a82891338f9bA80E2D6970FddA79D1eb0daE"

	r0, r1, err := GetUniV2PairReserves(poolAddr)
	if err != nil {
		return 0, nil
	}

	r0 = r0.Mul(r0, big.NewInt(1e18))
	r0 = r0.Div(r0, r1)
	priceWith1e18, _ := r0.Float64()

	return priceWith1e18 / 1e18, nil
}

func GetUniPoolType(poolAddr string) (int, error) {
	// uni v2 check
	_, err := getSingleProp(poolAddr, "kLast", getClient(), nil)
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

func GetUniV3PairFee(poolAddr string) (*big.Int, error) {
	fee, err := getSingleProp(poolAddr, "fee", getClient(), nil)
	if err != nil {
		utils.Warnf("[ GetUniV3PairFee ] get prop of pool(%v) error: %v", poolAddr, err)
		return nil, err
	}

	return fee.(*big.Int), nil
}

func TEST_POOL() {
	// tokenIn := "0x6a8C648C7635B50836285fD02ba5482d9526DEc0"
	// tokenOut := "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"

	// fee := big.NewInt(3000)
	// amountIn, _ := big.NewInt(0).SetString("2375067830253394", 10)

	// v3, err := GetUniV3SwapAmountOut(tokenIn, tokenOut, fee, amountIn)

	// fmt.Printf("[ TEST_POOL ] v3: %v, err: %v\n\n", v3, err)

	price, err := GetBnbPriceByBsc(nil)
	utils.Infof("price: %v, err: %v", price, err)

}
