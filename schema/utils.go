package schema

import (
	"context"
	"log"
	"sfilter/config"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CheckSwapEvent(topics []common.Hash) int {
	if isUniswapSwapV2Event(topics) {
		return SWAP_EVENT_UNISWAPV2_LIKE
	}

	if isUniswapSwapV3Event(topics) {
		return SWAP_EVENT_UNISWAPV3_LIKE
	}

	return SWAP_EVENT_UNKNOWN
}

func isUniswapSwapV2Event(topics []common.Hash) bool {
	if len(topics) == 3 {
		return strings.EqualFold(topics[0].String(), "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822")
	}

	return false
}

func isUniswapSwapV3Event(topics []common.Hash) bool {
	if len(topics) == 3 {
		return strings.EqualFold(topics[0].String(), "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67")
	}

	return false
}

func InitTables(mongodb *mongo.Client) {
	doInitTable(config.SwapTableName, SwapIndexModel, mongodb)
	doInitTable(config.BlockProceededTableName, BlockProceededIndexModel, mongodb)

	doInitTable(config.TokenTableName, TokenIndexModel, mongodb)
	doInitTable(config.PairTableName, PairIndexModel, mongodb)
}

func doInitTable(collectionName string, index []mongo.IndexModel, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(collectionName)

	filter := bson.M{"name": collectionName}
	cols, err := collection.Database().ListCollectionNames(context.Background(), filter)
	if err != nil {
		log.Fatal("[ InitTables] ListCollectionNames err: ", err)
	}

	if len(cols) == 0 {
		// 说明是新表, 则创建索引
		_, err = collection.Indexes().CreateMany(context.Background(), index)
		if err != nil {
			log.Fatalf("[ InitTables ] collection.Indexes().CreateMany error. name: % v, err: %v\n", collectionName, err)
			return
		}

		log.Printf("[ InitTables ] collection.Indexes().CreateMany for table: %v success\n", collectionName)
	} else {
		log.Printf("[ InitTables ] table exist, pass... collections: %v\n", cols)
	}
}
