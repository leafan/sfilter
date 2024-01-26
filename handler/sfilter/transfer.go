package handler

import (
	"fmt"
	"math"
	"math/big"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/token"
	"sfilter/services/transfer"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleTransfer(block *schema.Block, mongodb *mongo.Client) []*schema.Transfer {
	var transferSlices []*schema.Transfer

	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					transfer := parseTransferEvent(block, _log)
					if transfer != nil {
						// 保存到slice, 方便到时候修改在保存
						transferSlices = append(transferSlices, transfer)
					}
				}
			}

		}
	}

	return transferSlices
}

func UpsertTransferToDB(transfers []*schema.Transfer, swaps []*schema.Swap, mongodb *mongo.Client) {
	// 本区块内有过交易的 token 的 price 保存
	mainTokenPriceMap := make(map[string]float64)

	// 记录本区块内所有的txHash
	txhashMap := make(map[string]bool)

	// 先获取本区块交易中的价格
	for _, _swap := range swaps {
		mainTokenPriceMap[_swap.MainToken] = _swap.PriceInUsd
		txhashMap[_swap.TxHash] = true
	}

	for _, _transfer := range transfers {
		// 更新 usd value
		// 如果本区块的swap有交易过, 则直接update
		price, ok := mainTokenPriceMap[_transfer.Token]
		if ok && price > 0 {
			_transfer.TransferValueInUsd = _transfer.Amount * price
		} else {
			// 否则调用链上价格数据update
			_token, err := token.GetTokenInfo(_transfer.Token, mongodb)
			if err == nil {
				_transfer.TransferValueInUsd = _transfer.Amount * _token.PriceInUsd
			}
		}

		// 更新是否是swap类型的transfer
		_transfer.TransferType = schema.TRANSFER_EVENT_TRANSFER

		// 如果该hash里面也有swap交易, 直接认为是swap类型
		_, ok = txhashMap[_transfer.TxHash]
		if ok {
			_transfer.TransferType = schema.TRANSFER_EVENT_SWAP
		}

		transfer.SaveTransferEvent(_transfer, mongodb)
	}
}

func parseTransferEvent(block *schema.Block, l *types.Log) *schema.Transfer {
	var transfer *schema.Transfer

	if len(l.Topics) == 3 && strings.EqualFold(l.Topics[0].String(), "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef") {
		transfer = &schema.Transfer{
			Token: l.Address.String(),

			From: common.HexToAddress(l.Topics[1].Hex()).String(),
			To:   common.HexToAddress(l.Topics[2].Hex()).String(),

			AmountBigInt: new(big.Int).SetBytes(l.Data[:32]),

			BlockNo:  l.BlockNumber,
			TxHash:   l.TxHash.String(),
			Position: l.Index,

			Timestamp: time.Unix(int64(block.Block.Time()), 0),
		}

		// 获取token
		token, err := chain.GetTokenInfoForRead(transfer.Token)
		if err == nil {
			transfer.TokenSymbol = token.Symbol

			if token.Decimal <= 9 { // 没到丢精度的程度
				amount, _ := transfer.AmountBigInt.Float64()
				transfer.Amount = amount / math.Pow10(int(token.Decimal))
			} else {
				// 如果某个token少于1e9, 那可以当0看了
				tmpBig := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token.Decimal)-9), nil)
				tmpBig = tmpBig.Div(transfer.AmountBigInt, tmpBig)

				amount, _ := tmpBig.Float64()
				transfer.Amount = amount / 1e9
			}
		}

		transfer.LogIndexWithTx = fmt.Sprintf("%s_%d", transfer.TxHash, transfer.Position)
	}

	return transfer
}
