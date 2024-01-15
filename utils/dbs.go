package utils

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

func DoInitTable(database, collectionName string, index []mongo.IndexModel, mongodb *mongo.Client) {
	collection := mongodb.Database(database).Collection(collectionName)

	filter := bson.M{"name": collectionName}
	cols, err := collection.Database().ListCollectionNames(context.Background(), filter)
	if err != nil {
		Fatalf("[ InitTables] ListCollectionNames err: %v", err)
	}

	if len(cols) == 0 {
		// 说明是新表, 则创建索引
		_, err = collection.Indexes().CreateMany(context.Background(), index)
		if err != nil {
			Fatalf("[ InitTables ] collection.Indexes().CreateMany error. name: % v, err: %v\n", collectionName, err)
			return
		}

		Infof("[ InitTables ] collection.Indexes().CreateMany for table: %v success\n", collectionName)
	} else {
		Warnf("[ InitTables ] table exist, pass... collections: %v", cols)
	}
}
