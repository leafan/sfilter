package block

import (
	"context"
	"time"

	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func IsBlockProceeded(blkNo int64, mongodb *mongo.Client) bool {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	filter := bson.M{"blockNo": blkNo}

	var result bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return false
	}

	if len(result) == 0 {
		return false
	}

	return true
}

func SetBlockProceeded(bps *schema.BlockProceeded, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	bps.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), bps)
	if err != nil {
		utils.Warnf("[ SetBlockProceeded ] InsertOne error: %v, block no: %v\n", err, bps.BlockNo)
		return
	}

}

func SetUnProceeded(blockNo int64, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	filter := bson.M{"blockNo": blockNo}

	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		utils.Errorf("[ SetUnProceeded ] set failed, error: %v, blockNo: %v, ", err, blockNo)
	}

	utils.Infof("[ SetUnProceeded ] set success")
}
