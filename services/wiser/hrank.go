package wiser

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveTopRanks(ranks []interface{}, mongodb *mongo.Client) error {
	if len(ranks) <= 0 {
		return nil
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.HotPairRankTableName)

	session, err := mongodb.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	// 执行事务操作
	err = mongo.WithSession(context.Background(), session, func(sessionContext mongo.SessionContext) error {
		_, err := collection.InsertMany(sessionContext, ranks)
		return err
	})

	return err
}

func GetHotRankToMap(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Client) (schema.HRankMap, error) {
	hmap := make(schema.HRankMap)

	ranks, err := GetHotRanks(findOpt, filter, mongodb)
	if err != nil {
		return nil, err
	}

	// 将ranks存入map
	for _, rank := range ranks {
		if _, ok := hmap[rank.PairAddress]; !ok {
			hmap[rank.PairAddress] = make([]schema.HotPairRank, 0)
		}
		hmap[rank.PairAddress] = append(hmap[rank.PairAddress], rank)
	}

	// 再将每一个pair里面的多个hrank按createdAt升序排列, 表明时间从旧到新
	for _, hrank := range hmap {
		sort.Slice(hrank, func(i, j int) bool {
			return hrank[i].CreatedAt.Before(hrank[j].CreatedAt)
		})
	}

	return hmap, nil
}

func GetHotRanks(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Client) ([]schema.HotPairRank, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.HotPairRankTableName)

	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	var result []schema.HotPairRank
	cursor, err := collection.Find(ctx, filter, findOpt)
	if err != nil {
		return result, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &result)
	return result, err
}
