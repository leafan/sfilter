package pair

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var pairLock sync.Mutex

func UpdatePoolInfo(address string, traderInfo *schema.InfoOnPools, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	info := struct {
		schema.InfoOnPools `bson:",inline"`
		UpdatedAt          time.Time `bson:"updatedAt"`
	}{
		InfoOnPools: *traderInfo,
		UpdatedAt:   time.Now(),
	}
	filter := bson.D{{Key: "address", Value: address}}

	update := bson.D{
		{Key: "$set", Value: info},
	}

	pairLock.Lock()
	defer pairLock.Unlock()

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Warnf("[ UpdatePoolInfo ] failed. pair: %v, err: %v\n", address, err)
	}
}

// 如果存在就更新, 不存在就插入
func UpSertOnChainInfo(address string, infoOnChain *schema.InfoOnChain, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	info := struct {
		schema.InfoOnChain `bson:",inline"`
		UpdatedAt          time.Time `bson:"updatedAt"`
	}{
		InfoOnChain: *infoOnChain,
		UpdatedAt:   time.Now(),
	}

	filter := bson.D{{Key: "address", Value: address}}

	update := bson.D{
		{Key: "$set", Value: info},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "createdAt", Value: time.Now()},
		}},
	}

	opt := options.Update().SetUpsert(true) // 执行更新操作，设置upsert为true

	pairLock.Lock()
	defer pairLock.Unlock()

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ UpSertOnChainInfo ] failed. pair: %v, err: %v\n", address, err)
	}
}

func UpSertPairCreatedInfo(pair *schema.Pair, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	info := struct {
		schema.InfoOnChain       `bson:",inline"`
		schema.InfoOnPairCreated `bson:",inline"`
		UpdatedAt                time.Time `bson:"updatedAt"`
	}{
		InfoOnChain:       pair.InfoOnChain,
		InfoOnPairCreated: pair.InfoOnPairCreated,
		UpdatedAt:         time.Now(),
	}

	filter := bson.D{{Key: "address", Value: pair.Address}}

	update := bson.D{
		{Key: "$set", Value: info},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "createdAt", Value: time.Now()},
		}},
	}

	opt := options.Update().SetUpsert(true) // 执行更新操作，设置upsert为true

	pairLock.Lock()
	defer pairLock.Unlock()

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		utils.Warnf("[ UpSertPairCreatedInfo ] failed. pair: %v, err: %v\n", pair.Address, err)
	}
}

func UpdateTradeInfo(pair *schema.Pair, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	// 更新时间
	pair.TradeInfoUpdatedAt = time.Now()

	info := struct {
		schema.TradeInfoForPair `bson:",inline"`
		UpdatedAt               time.Time `bson:"updatedAt"`
	}{
		TradeInfoForPair: pair.TradeInfoForPair,
		UpdatedAt:        time.Now(),
	}
	filter := bson.D{{Key: "address", Value: pair.Address}}

	update := bson.D{
		{Key: "$set", Value: info},
	}

	pairLock.Lock()
	defer pairLock.Unlock()

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Warnf("[ UpdateTradeInfo ] failed. pair: %v, err: %v\n", pair.Address, err)
	}
}

// 更新整个pair
func UpsertPair(pair *schema.Pair, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)
	filter := bson.D{{Key: "address", Value: pair.Address}}

	update := bson.D{
		{Key: "$set", Value: pair},
	}

	pairLock.Lock()
	defer pairLock.Unlock()

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Warnf("[ UpsertPair ] failed. pair: %v, err: %v\n", pair.Address, err)
	}
}
