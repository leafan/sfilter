package handler

import (
	"log"
	"math/big"
	"sfilter/schema"
	"sfilter/utils"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
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
			log.Printf("[ parseUniV2AddLiquidity ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
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
			log.Printf("[ parseLiquidityEvent ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
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
			log.Printf("[ parseUniV2RemoveLiquidity ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
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
			log.Printf("[ parseUniV3RemoveLiquidity ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
		} else {
			event.Operator = sender.String()
		}

		event.Amount0 = new(big.Int).SetBytes(l.Data[32:64]).String()
		event.Amount1 = new(big.Int).SetBytes(l.Data[64:96]).String()
	}

	return event
}

func updateLiquidityAmount(event *schema.LiquidityEvent, _pair *schema.Pair, block *schema.Block) {
	event.PairName = _pair.PairName

	amount := utils.CalculateVolumeInUsd(_pair.Token0, utils.GetBigIntOrZero(event.Amount0), _pair.Decimal0, block.EthPrice)
	amount += utils.CalculateVolumeInUsd(_pair.Token1, utils.GetBigIntOrZero(event.Amount1), _pair.Decimal1, block.EthPrice)

	event.AmountInUsd = amount
}
