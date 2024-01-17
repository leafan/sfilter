package pair

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

func GetHotPairs(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]*schema.Pair, int64, error) {
	collection := mongodb.Collection(config.PairTableName)

	var result []*schema.Pair
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		utils.Warnf("[ GetHotPairs ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}

func GetPairMap(pageSize int64, mongodb *mongo.Database) (schema.PairMap, error) {
	pairMap := make(schema.PairMap)

	collection := mongodb.Collection(config.PairTableName)
	filter := bson.M{}

	page := int64(1)
	skip := (page - 1) * pageSize

	for {
		cursor, err := collection.Find(context.Background(), filter, options.Find().SetLimit(pageSize).SetSkip(skip))
		if err != nil {
			return pairMap, err
		}

		count := 0
		for cursor.Next(context.Background()) {
			var pair schema.Pair
			if err := cursor.Decode(&pair); err != nil {
				cursor.Close(context.Background())
				return pairMap, err
			}
			count++

			//  针对某一笔swap, 处理出对应数据
			pairMap[pair.Address] = pair
		}

		if err := cursor.Err(); err != nil {
			cursor.Close(context.Background())
			return pairMap, err
		}

		if count < int(pageSize) {
			// reached the end of data
			break
		}
		skip += pageSize
	}

	return pairMap, nil
}
