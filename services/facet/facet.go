package facet

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

func SaveFacet(fct *schema.FacetModel, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.FacetTableName)

	_, err := collection.InsertOne(context.Background(), fct)
	if err != nil {
		utils.Warnf("[ SaveFacet ] InsertOne error: %v, fct: %v, hash: %v\n", err, fct, fct.TxHash)
	}
}

func GetFacets(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.FacetModel, int64, error) {
	collection := mongodb.Collection(config.FacetTableName)

	var result []schema.FacetModel
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
		utils.Warnf("[ GetFacets ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}
