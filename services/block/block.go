package block

import (
	"context"
	"log"
	"time"

	"sfilter/config"
	"sfilter/schema"

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

func SaveBlockProceeded(bps *schema.BlockProceeded, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	bps.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), bps)
	if err != nil {
		log.Printf("[ SaveBlockProceeded ] InsertOne error: %v, block no: %v\n", err, bps.BlockNo)
		return
	}

}
