package handler

import (
	"fmt"
	"math/big"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/transfer"
	"strings"
	"time"
    "math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleTransfer(block *schema.Block, mongodb *mongo.Client) schema.TxTokenTransfersMap {
	maps := make(schema.TxTokenTransfersMap)

	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					transfer := parseTransferEvent(block, _log)
					if transfer != nil {
						// log.Printf("[ HandleTransfer ] token addr: %v, txWithIndex: %v\n\n", transfer.Token, transfer.LogIndexWithTx)

						handleTransfer(transfer, mongodb)

						// 保存进 map, 方便swap的时候查找
						key := fmt.Sprintf("%v_%v", tx.OriginTx.Hash().String(), transfer.Token)
						maps[key] = append(maps[key], transfer)
					}
				}
			}

		}
	}

	return maps
}

func handleTransfer(_transfer *schema.Transfer, mongodb *mongo.Client) {
	transfer.SaveTransferEvent(_transfer, mongodb)
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
