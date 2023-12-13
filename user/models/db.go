package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"sfilter/utils"
)

func initTable(collection *mongo.Collection, collectionName string, index []mongo.IndexModel, mongodb *mongo.Client) {
	filter := bson.M{"name": collectionName}
	cols, err := collection.Database().ListCollectionNames(context.Background(), filter)
	if err != nil {
		utils.Fatalf("[ InitTables] ListCollectionNames err: %v", err)
	}

	if len(cols) == 0 {
		// 说明是新表, 则创建索引
		_, err = collection.Indexes().CreateMany(context.Background(), index)
		if err != nil {
			utils.Fatalf("[ InitTables ] collection.Indexes().CreateMany error. name: % v, err: %v", collectionName, err)
			return
		}

		utils.Infof("[ InitTables ] collection.Indexes().CreateMany for table: %v success", collectionName)
	} else {
		utils.Warnf("[ InitTables ] table exist, pass... collections: %v", cols)
	}
}
