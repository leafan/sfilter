package handler

import (
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/pair"
	"sfilter/utils"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandlePairLogic(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					handlePairCreated(block, _log, mongodb)
				}
			}
		}
	}
}

func handlePairCreated(block *schema.Block, _log *types.Log, mongodb *mongo.Client) {
	_pair := parsePairCreatedEvent(_log)
	if _pair != nil {
		utils.Debugf("[ handlePairCreated ] pair: %v, tx: %v", _pair, _log.TxHash)

		block.PairCreatedNum++

		updatePairTokenInfo(_pair, mongodb)

		// 插入或更新pair
		pair.UpSertPairCreatedInfo(_pair, mongodb)
	}
}

func updatePairTokenInfo(_pair *schema.Pair, mongodb *mongo.Client) {
	token0, err0 := chain.GetTokenInfo(_pair.Token0, mongodb)
	token1, err1 := chain.GetTokenInfo(_pair.Token1, mongodb)
	if err0 != nil || err1 != nil {
		// pair create的时候, 一般不会出错..
		utils.Warnf("[ updatePairTokenInfo ] getTokenInfo error. token0: %v, err0: %v, token1: %v, err1: %v\n", _pair.Token0, err0, _pair.Token1, err1)

		_pair.Decimal0 = 0
		_pair.Decimal1 = 0
		_pair.PairName = "Unknown/Unknown"

		return
	}

	_pair.Decimal0 = token0.Decimal
	_pair.Decimal1 = token1.Decimal

	pair.GeneratePairName(_pair, token0, token1)
}
