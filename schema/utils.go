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
		return strings.EqualFold(topics[0].String(), "0xbeee1e6e7fe307ddcf84b0a16137a4430ad5e2480fc4f4a8e250ab56ccd7630d")
	}

	return false
}

func InitTables(mongodb *mongo.Client) {
	initSwapTable(mongodb)
	initBlockProceededTable(mongodb)
}

func initBlockProceededTable(mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	filter := bson.M{"name": config.BlockProceededTableName}
	_, err := collection.Database().ListCollectionNames(context.Background(), filter)
	if err != nil {
		log.Fatal("[ InitTables] ListCollectionNames err: ", err)
	}

	if err == mongo.ErrNilDocument {
		// 说明是新表, 则创建索引
		_, err = collection.Indexes().CreateMany(context.Background(), SwapIndexModel)
		if err != nil {
			log.Fatal("[ InitTables ] collection.Indexes().CreateMany error: ", err)
			return
		}

		log.Printf("[ InitTables ] collection.Indexes().CreateMany success\n")
	} else {
		log.Printf("[ InitTables ] BlockProceeded table exist, pass...")
	}
}

// 如果不存在表, 则创建索引
func initSwapTable(mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.SwapTableName)

	filter := bson.M{"name": config.SwapTableName}
	_, err := collection.Database().ListCollectionNames(context.Background(), filter)
	if err != nil {
		log.Fatal("[ InitTables] ListCollectionNames err: ", err)
	}

	if err == mongo.ErrNilDocument {
		// 说明是新表, 则创建索引
		_, err = collection.Indexes().CreateMany(context.Background(), SwapIndexModel)
		if err != nil {
			log.Fatal("[ InitTables ] collection.Indexes().CreateMany error: ", err)
			return
		}

		log.Printf("[ InitTables ] collection.Indexes().CreateMany success\n")
	} else {
		log.Printf("[ InitTables ] swap table exist, pass...")
	}
}

// debug
func PrintOneBlock(block *Block) {
	log.Println("Number        : ", block.Block.Number())
	log.Println("Hash            : ", block.Block.Hash().Hex())

	for i, tx := range block.Transactions {
		log.Printf("tx %d hash: %v\n", i, tx.OriginTx.Hash())
		log.Printf("tx %d log count: %d \n", i, len(tx.Receipt.Logs))

		log.Printf("tx %d log[0]: %v\n", i, tx.Receipt.Logs[0])
	}
}
