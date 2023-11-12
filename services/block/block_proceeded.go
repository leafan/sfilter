package block

import (
	"context"
	"log"

	"sfilter/config"
	"sfilter/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SaveBlockProceeded(bps *schema.BlockProceeded, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	_, err := collection.InsertOne(context.Background(), bps)
	if err != nil {
		log.Printf("[ SaveBlockProceeded ] InsertOne error: %v, block no: %v\n", err, bps.BlockNo)
		return
	}

}

func IsBlockProceeded(blkNo int64, mongodb *mongo.Client) bool {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BlockProceededTableName)

	filter := bson.M{"blockno": blkNo}

	var result bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		// log.Printf("[ isBlockProceeded ] Find block %v failed. err: %v\n", blkNo, err)
		return false
	}

	if len(result) == 0 {
		// log.Println("[ isBlockProceeded ] FindOne result 0...")
		return false
	}

	// log.Println("[ isBlockProceeded ] FindOne result 1...")
	return true
}
