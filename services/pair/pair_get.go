package pair

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetNewPairs(mongodb *mongo.Client) ([]schema.Pair, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

    limit := int64(5)
    page := int64(2)
    skip := int64(page * limit - limit)
    
    options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "firstAddPoolTime", Value: -1}})

	var result []schema.Pair
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{}, options)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("[ GetNewPairs ] Find error: %v\n", err)
		}

		return result, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &result)
	return result, err
}