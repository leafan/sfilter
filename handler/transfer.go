package handler

import (
	"fmt"
	"log"
	"math/big"
	"sfilter/schema"
	"sfilter/services/transfer"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleTransfer(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					transfer := parseTransferEvent(block, _log)
					if transfer != nil {
						log.Printf("[ HandleTransfer ] token addr: %v, txWithIndex: %v\n\n", transfer.Token, transfer.LogIndexWithTx)

						handleTransfer(transfer, mongodb)
					}
				}
			}

		}
	}
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

			Amount: new(big.Int).SetBytes(l.Data[:32]).String(),

			BlockNo:  l.BlockNumber,
			TxHash:   l.TxHash.String(),
			Position: l.Index,

			Timestamp: time.Unix(int64(block.Block.Time()), 0),
		}

		transfer.LogIndexWithTx = fmt.Sprintf("%s_%d", transfer.TxHash, transfer.Position)

	}

	return transfer
}
