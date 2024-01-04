package transfer

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveTransferEvent(_transfer *schema.Transfer, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TransferTableName)

	_transfer.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), _transfer)
	if err != nil {
		// utils.Warnf("[ SaveTransferEvent ] InsertOne error: %v, trasnfer: %v, LogIndexWithTx: %v\n", err, _transfer, _transfer.LogIndexWithTx)
	}
}

func GetTransferEvents(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.Transfer, int64, error) {
	collection := mongodb.Collection(config.TransferTableName)

	var result []schema.Transfer
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, findOpt)
	if err != nil {
		return result, 0, err
	}
	defer cursor.Close(ctx)

	// 限制 count 上限, 否则会卡死, 查询太久的也没有意义
	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	totalCount, err := collection.CountDocuments(ctx, filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetTransferEvents ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}
