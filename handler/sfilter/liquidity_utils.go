package handler

// !!!!!
// int64 很容易被 1e18 溢出	！

import (
	"math/big"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/pair"
	"sfilter/utils"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// v2 add
// refer: https://cn.etherscan.com/tx/0x61c68478289b15b67e22576946ad1f78ed74d07c6bf07f23e71470c02b578fc1#eventlog
func parseUniV2AddLiquidity(l *types.Log, tx *schema.Transaction) *schema.LiquidityEvent {
	var event *schema.LiquidityEvent

	if len(l.Topics) == 2 && strings.EqualFold(l.Topics[0].String(), "0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f") && len(l.Data) == 64 {
		event = &schema.LiquidityEvent{
			PoolAddress: l.Address.String(),
			Direction:   schema.DIRECTION_BUY_OR_ADD, // add
			Type:        schema.SWAP_EVENT_UNISWAPV2_LIKE,
		}

		sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
		if err != nil {
			utils.Warnf("[ parseUniV2AddLiquidity ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
		} else {
			event.Operator = sender.String()
		}

		event.Amount0 = new(big.Int).SetBytes(l.Data[0:32]).String()
		event.Amount1 = new(big.Int).SetBytes(l.Data[32:64]).String()
	}

	return event
}

// v3 add
func parseUniV3AddLiquidity(l *types.Log, tx *schema.Transaction) *schema.LiquidityEvent {
	var event *schema.LiquidityEvent

	// uni v3 的 Mint.. 第一版本监控的 IncreaseLiquidity 函数, 需要根据tokenId反查..
	// refer: https://etherscan.io/tx/0x414eb36cfea5fb34e068a79196480e20caf021fb0921cf148f5aa0967faffe1d#eventlog
	if len(l.Topics) == 4 && strings.EqualFold(l.Topics[0].String(), "0x7a53080ba414158be7ec69b987b5fb7d07dee101fe85488f0853ae16239d0bde") && len(l.Data) == 128 {
		event = &schema.LiquidityEvent{
			PoolAddress: l.Address.String(),
			Direction:   schema.DIRECTION_BUY_OR_ADD, // add
			Type:        schema.SWAP_EVENT_UNISWAPV3_LIKE,
		}

		sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
		if err != nil {
			utils.Warnf("[ parseLiquidityEvent ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
		} else {
			event.Operator = sender.String()
		}

		event.Amount0 = new(big.Int).SetBytes(l.Data[64:96]).String()
		event.Amount1 = new(big.Int).SetBytes(l.Data[96:128]).String()
	}

	return event
}

// v2 remove, 只有一个burn函数
// refer: https://etherscan.io/tx/0xf083a1841761ff17221f870ff0f84a46a35388dbe05bc2c93a6c284f257d846a#eventlog
func parseUniV2RemoveLiquidity(l *types.Log, tx *schema.Transaction) *schema.LiquidityEvent {
	var event *schema.LiquidityEvent

	if len(l.Topics) == 3 && strings.EqualFold(l.Topics[0].String(), "0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496") && len(l.Data) == 64 {
		event = &schema.LiquidityEvent{
			PoolAddress: l.Address.String(),
			Direction:   schema.DIRECTION_SELL_OR_DECREASE,
			Type:        schema.SWAP_EVENT_UNISWAPV2_LIKE,
		}

		sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
		if err != nil {
			utils.Warnf("[ parseUniV2RemoveLiquidity ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
		} else {
			event.Operator = sender.String()
		}

		event.Amount0 = new(big.Int).SetBytes(l.Data[0:32]).String()
		event.Amount1 = new(big.Int).SetBytes(l.Data[32:64]).String()
	}

	return event
}

// v3 remove
// refer: https://etherscan.io/tx/0x74aa183831ed9dd9672807c6768b570e11e3c9700d3ad151e315bfc8000a83ee#eventlog
func parseUniV3RemoveLiquidity(l *types.Log, tx *schema.Transaction) *schema.LiquidityEvent {
	var event *schema.LiquidityEvent

	// burn
	if len(l.Topics) == 4 && strings.EqualFold(l.Topics[0].String(), "0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c") && len(l.Data) == 96 {
		event = &schema.LiquidityEvent{
			PoolAddress: l.Address.String(),
			Direction:   schema.DIRECTION_SELL_OR_DECREASE,
			Type:        schema.SWAP_EVENT_UNISWAPV3_LIKE,
		}

		sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
		if err != nil {
			utils.Warnf("[ parseUniV3RemoveLiquidity ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
		} else {
			event.Operator = sender.String()
		}

		event.Amount0 = new(big.Int).SetBytes(l.Data[32:64]).String()
		event.Amount1 = new(big.Int).SetBytes(l.Data[64:96]).String()
	}

	return event
}

func updateLiquidityEventValue(event *schema.LiquidityEvent, _pair *schema.Pair, block *schema.Block) {
	event.PairName = _pair.PairName

	var amount, amount0, amount1 float64

	amount0 = utils.CalculateVolumeInUsd(_pair.Token0, utils.GetBigFloatOrZero(event.Amount0), _pair.Decimal0, block.EthPrice)

	amount1 = utils.CalculateVolumeInUsd(_pair.Token1, utils.GetBigFloatOrZero(event.Amount1), _pair.Decimal1, block.EthPrice)

	amount = amount0 + amount1

	// 如果两方都不为0, 且一方usd为0, 则直接乘以2
	if event.Type == schema.SWAP_EVENT_UNISWAPV2_LIKE && (amount0 == 0 || amount1 == 0) {
		if event.Amount0 != "" && event.Amount1 != "" && event.Amount0 != "0" && event.Amount1 != "0" {
			amount *= 2
		}
	}

	event.AmountInUsd = amount
}

// 去链上获取流动性池子大小
// 直接获取token0及token1的balance, 再确认价值币
func UpdatePoolLiquidity(_pair *schema.Pair, mongodb *mongo.Client, block *schema.Block) {
	token0BalanceInt, err0 := chain.BalanceOf(_pair.Address, _pair.Token0)
	token1BalanceInt, err1 := chain.BalanceOf(_pair.Address, _pair.Token1)
	if err0 != nil || err1 != nil {
		utils.Warnf("[ updatePoolLiquidity ] get balance err0: %v, err1: %v\n", err0, err1)
		return
	}

	token0Balance := utils.GetBigFloatOrZero(token0BalanceInt.String())
	token1Balance := utils.GetBigFloatOrZero(token1BalanceInt.String())

	amount0 := utils.CalculateVolumeInUsd(_pair.Token0, token0Balance, _pair.Decimal0, block.EthPrice)
	amount1 := utils.CalculateVolumeInUsd(_pair.Token1, token1Balance, _pair.Decimal1, block.EthPrice)

	// utils.Warnf("[ updatePoolLiquidity ] before.. block: %v, pair: %v liquidity now.. amount0: %v, balance0: %v,  amount1: %v, balance1: %v", block.BlockNo, _pair.Address, amount0, token0Balance, amount1, token1Balance)

	// 计算pool价值币value直接相加即可
	_pair.ValueCoinLiquidity = amount0 + amount1

	// 计算 liquidity
	// 1. 如果双方为0, 则为0
	// 2. 如果双方均不为0, 则直接相加
	// 3. 如果某一方为0, 另一方不为0, 则根据价格计算为0方价值
	if (amount0 != 0 && amount1 == 0) || (amount0 == 0 && amount1 != 0) {
		// 因为price的计算时, 已经针对价值币进行了分类
		// 因此到这里的时候, 如果一方为0, 那肯定价格就是0价值方价格
		// 所以直接乘以价格即可
		if _pair.Price > 0 {
			token0Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(_pair.Decimal0)), nil)
			token1Exponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(_pair.Decimal1)), nil)

			if amount1 == 0 { // token0是价值币, token1为屌丝币
				// int64 很容易被 1e18 溢出	！当price大于100时就溢出了!
				// 所以换成 big.Float....
				amount1BigInt := token1Balance.Mul(token1Balance, big.NewFloat(_pair.Price*1e18)) // 防止价格过低的屌丝币, 乘上一个基数

				amount1BigInt = amount1BigInt.Quo(amount1BigInt, new(big.Float).SetInt(token1Exponent))

				// 此时的amount1BigInt表示该屌丝币兑换成对端价值币后的数量
				// 所以还需要乘以对端的decimal才表示真实的 amount
				amount1BigInt = amount1BigInt.Mul(amount1BigInt, new(big.Float).SetInt(token0Exponent))

				// 再除以基数
				amount1BigInt = amount1BigInt.Quo(amount1BigInt, big.NewFloat(1e18))

				amount1 = utils.CalculateVolumeInUsd(_pair.Token0, amount1BigInt, _pair.Decimal0, block.EthPrice)
			} else { // 反之
				amount0BigInt := token0Balance.Mul(token0Balance, big.NewFloat(_pair.Price*1e18))
				amount0BigInt = amount0BigInt.Quo(amount0BigInt, new(big.Float).SetInt(token0Exponent))
				amount0BigInt = amount0BigInt.Mul(amount0BigInt, new(big.Float).SetInt(token1Exponent))

				amount0BigInt = amount0BigInt.Quo(amount0BigInt, big.NewFloat(1e18))
				amount0 = utils.CalculateVolumeInUsd(_pair.Token1, amount0BigInt, _pair.Decimal1, block.EthPrice)
			}
		}
	}
	_pair.LiquidityInUsd = amount0 + amount1

	_pair.Token0UsdValue = amount0
	_pair.Token1UsdValue = amount1

	// update pair liquidity info
	// utils.Infof("[ updatePoolLiquidity ] update pair: %v liquidity now.. amount0: %v, amount1: %v", _pair.Address, amount0, amount1)
	pair.UpSertOnChainInfo(_pair.Address, &_pair.InfoOnChain, mongodb)
}
