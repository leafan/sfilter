package schema

import (
	"context"
	"sfilter/config"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitTables(mongodb *mongo.Client) {
	doInitTable(config.SwapTableName, SwapIndexModel, mongodb)
	doInitTable(config.BlockProceededTableName, BlockProceededIndexModel, mongodb)

	doInitTable(config.TokenTableName, TokenIndexModel, mongodb)
	doInitTable(config.PairTableName, PairIndexModel, mongodb)
	doInitTable(config.LiquidityEventTableName, LiquidityEventIndexModel, mongodb)

	doInitTable(config.Kline1MinTableName, Kline1MinIndexModel, mongodb)
	doInitTable(config.Kline1HourTableName, Kline1HourIndexModel, mongodb)

	doInitTable(config.TransferTableName, TransferIndexModel, mongodb)

	doInitTable(config.ConfigTableName, ConfigIndexModel, mongodb)
	doInitTable(config.GlobalTrendTableName, GlobalTrendIndexModel, mongodb)

}

func doInitTable(collectionName string, index []mongo.IndexModel, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(collectionName)

	filter := bson.M{"name": collectionName}
	cols, err := collection.Database().ListCollectionNames(context.Background(), filter)
	if err != nil {
		utils.Fatalf("[ InitTables] ListCollectionNames err: ", err)
	}

	if len(cols) == 0 {
		// 说明是新表, 则创建索引
		_, err = collection.Indexes().CreateMany(context.Background(), index)
		if err != nil {
			utils.Fatalf("[ InitTables ] collection.Indexes().CreateMany error. name: % v, err: %v\n", collectionName, err)
			return
		}

		utils.Infof("[ InitTables ] collection.Indexes().CreateMany for table: %v success\n", collectionName)
	} else {
		utils.Warnf("[ InitTables ] table exist, pass... collections: %v", cols)
	}
}
