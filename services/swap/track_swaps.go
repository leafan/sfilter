package swap

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveTrackSwap(uswap *schema.TrackSwap, mongodb *mongo.Client) error {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TrackSwapTableName)

	// 先查找是否已存在, 如果hash和user
	// todo

	uswap.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), uswap)
	if err != nil {
		utils.Warnf("[ SaveTrackSwap ] InsertOne error: %v, swap tx: %v\n", err, uswap.TxHash)
	}

	return err
}

func GetTrackSwaps(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.TrackSwap, int64, error) {
	collection := mongodb.Collection(config.TrackSwapTableName)

	var result []schema.TrackSwap
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, findOpt)
	if err != nil {
		return result, 0, err
	}
	defer cursor.Close(ctx)

	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	totalCount, err := collection.CountDocuments(ctx, filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetTrackSwaps ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}

func GetUserTrackSwapCount(username string, mongodb *mongo.Client) (int64, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TrackSwapTableName)

	filter := bson.M{}
	filter["username"] = username

	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()
	totalCount, err := collection.CountDocuments(ctx, filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetUserTrackSwapCount ] Count error: %v\n", err)
		return 0, err
	}

	return totalCount, nil
}

func DeleteOldEntries(username string, deleteCount int64, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TrackSwapTableName)

	cursor, err := collection.Find(
		context.Background(),
		bson.M{"username": username},
		options.Find().SetSort(bson.M{"swapTime": 1}).SetProjection(bson.M{"_id": 1}).SetLimit(deleteCount),
	)
	if err != nil {
		utils.Errorf("[ DeleteOldEntries ] find delete data error: %v", err)
		return
	}
	defer cursor.Close(context.Background())

	var deleteIDs []interface{}
	for cursor.Next(context.Background()) {
		var result bson.M

		err := cursor.Decode(&result)
		if err != nil {
			utils.Errorf("[ DeleteOldEntries ] cursor.Decode error: %v", err)
			return
		}

		id := result["_id"]
		deleteIDs = append(deleteIDs, id)
	}

	deleteFilter := bson.M{"_id": bson.M{"$in": deleteIDs}}
	_, err = collection.DeleteMany(context.Background(), deleteFilter)
	if err != nil {
		utils.Errorf("[ DeleteOldEntries ] DeleteMany error: %v", err)
		return
	}

	utils.Infof("[ DeleteOldEntries ] delete success, count: %v", deleteCount)
}
