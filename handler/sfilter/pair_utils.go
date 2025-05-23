package handler

import (
	"math/big"
	"sfilter/schema"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func parsePairCreatedEvent(l *types.Log) *schema.Pair {
	var pair *schema.Pair

	// uniswap v2 的 paircreated 事件
	// refer: https://etherscan.io/tx/0x0b8fd248b6ed148d87995d990bf5031da2398ae8370fae798959fa4ce7249804#eventlog
	if len(l.Topics) == 3 && strings.EqualFold(l.Topics[0].String(), "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9") {
		pair = &schema.Pair{
			InfoOnChain: schema.InfoOnChain{
				Address: common.BytesToAddress(l.Data[0:32]).String(),
				Token0:  common.HexToAddress(l.Topics[1].Hex()).String(),
				Token1:  common.HexToAddress(l.Topics[2].Hex()).String(),
			},

			InfoOnPairCreated: schema.InfoOnPairCreated{
				Type:               schema.SWAP_EVENT_UNISWAPV2_LIKE,
				PairCreatedBlockNo: l.BlockNumber,
				PairCreatedHash:    l.TxHash.String(),
				PairFee:            3000,
			},

			CreatedAt: time.Now(),
		}
	}

	if len(l.Topics) == 4 && strings.EqualFold(l.Topics[0].String(), "0x783cca1c0412dd0d695e784568c96da2e9c22ff989357a2e8b1d9b2b4e6b7118") {
		// uniswap v3
		// refer: https://etherscan.io/tx/0xb9bc7e088ea5cd41398ff4e6a50c725a1065b653f365aaab702b7dbd85a107cd#eventlog
		pair = &schema.Pair{
			InfoOnChain: schema.InfoOnChain{
				Address: common.BytesToAddress(l.Data[32:64]).String(),
				Token0:  common.HexToAddress(l.Topics[1].Hex()).String(),
				Token1:  common.HexToAddress(l.Topics[2].Hex()).String(),
			},

			InfoOnPairCreated: schema.InfoOnPairCreated{
				Type:               schema.SWAP_EVENT_UNISWAPV3_LIKE,
				PairCreatedBlockNo: l.BlockNumber,

				PairCreatedHash: l.TxHash.String(),
			},

			CreatedAt: time.Now(),
		}

		fee, _ := new(big.Int).SetString(l.Topics[3].Hex(), 0)
		pair.PairFee = fee.Int64()
	}

	return pair
}
