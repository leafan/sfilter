package swap

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/utils"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var kline1minLock sync.Mutex

func UpdateKlines(swap *schema.Swap, mongodb *mongo.Client) {
	update1MinKline(swap, mongodb)
	update1DayKline(swap, mongodb)
}

func update1MinKline(swap *schema.Swap, mongodb *mongo.Client) {
	// 防止同时写入的时候, 查询时覆盖
	// 简单实现, 实际上应该是给 行加读写锁, 这里直接给整个加锁了..

	if swap.Price == "" {
		log.Printf("[ update1MinKline ] wrong price. swap: %v, tx: %v\n", swap, swap.LogIndexWithTx)
		return
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1MinTableName)

	symbol := fmt.Sprintf("%v/%v", swap.Token0, swap.Token1)
	quoteToken := swap.Token1
	if swap.MainToken == swap.Token1 {
		symbol = fmt.Sprintf("%v/%v", swap.Token1, swap.Token0)
		quoteToken = swap.Token0
	}

	_time := time.Unix(swap.SwapTime, 0) // 以交易(区块)时间为准, 而不是当前时间
	key := fmt.Sprintf("%v_%v_%v", symbol, _time.Day(), _time.Hour())

	// log.Printf("[ update1MinKline ] come here key: %v, minute: %v, swap: %v\n", key, _time.Minute(), swap)

	filter := bson.M{"symbolDayHour": key}

	var kline schema.KLines1Min

	// 直接整体加锁吧, 没啥性能问题
	kline1minLock.Lock()
	defer kline1minLock.Unlock()

	err := collection.FindOne(context.Background(), filter).Decode(&kline)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("[ update1MinKline ] FindOne error: %v, swap tx: %v\n", err, swap.LogIndexWithTx)
		return
	}

	if kline.SymbolDayHour == "" {
		// 说明没有这条k线, 需要新建; 填充基础信息
		kline.Symbol = symbol
		kline.BaseToken = swap.MainToken
		kline.QuoteToken = quoteToken

		kline.SymbolDayHour = key

		kline.Timestamp = _time
	} else {
		// 非新柱子, 则要判断下是否为当前时间柱子, 还是老柱子
		// 跑到这里的时候, 两个时间戳的 Hour 一定一样, 此时如果相差超过1小时, 那肯定是新柱子
		timeDiff := _time.Sub(kline.Timestamp).Abs().Hours()
		if timeDiff > 1.0 {
			kline.Timestamp = _time
			kline.Kline = schema.KLinesForHour{} // 清空柱子
		}
	}

	// 跑到这里来的时候, 一定是本时间段的柱子
	// 如果不是, 则柱子已经被清空

	candisk := &kline.Kline[_time.Minute()] // 指针直接修改
	updateKLineWithNewData(candisk, swap)

	SaveUpsert1MinKline(&kline, mongodb)
}

func updateKLineWithNewData(kline *schema.KLine, swap *schema.Swap) {
	// log.Printf("[ updateKLineWithNewData ] before update.. price: %v, volume: %v, kline: %v\n", swap.Price, swap.AmountOfMainToken, kline)

	bigPrice, ok := new(big.Float).SetString(swap.Price)
	if !ok {
		log.Printf("[ updateKLineWithNewData ] wrong price. price: %v, tx: %v\n", swap.Price, swap.LogIndexWithTx)
		return
	}

	bigPrice = bigPrice.Quo(bigPrice, new(big.Float).SetFloat64(config.PriceBaseFactor))
	curPrice, _ := bigPrice.Float64()

	kline.ClosePrice = curPrice // 不管新老柱子, 先更新close

	if kline.UnixTime == 0 {
		// 新柱子
		kline.OpenPrice = curPrice

		kline.HighPrice = curPrice
		kline.LowPrice = curPrice
	} else {
		// 不是新柱子
		if curPrice > kline.HighPrice {
			kline.HighPrice = curPrice
		}

		if curPrice < kline.LowPrice {
			kline.LowPrice = curPrice
		}
	}

	kline.UnixTime = swap.SwapTime

	// volume啥都不用考虑, 直接加
	volume, ok := new(big.Int).SetString(swap.AmountOfMainToken, 10)
	if !ok {
		log.Printf("[ updateKLineWithNewData ] wrong volume. AmountOfMainToken: %v, tx: %v\n", swap.AmountOfMainToken, swap.LogIndexWithTx)
		// volume = big.NewInt(0)
		return // volume都为0了, 没必要计算tx啥的了
	}
	kline.Volume = volume.Add(volume, utils.GetBigIntOrZero(kline.Volume)).String()

	// 更新 deepeye info
	kline.TxNum++

	oldUsdVolume := utils.GetBigIntOrZero(kline.VolumeInUsd)
	volumeInUsd := utils.GetBigIntOrZero(swap.VolumeInUsd)
	kline.VolumeInUsd = oldUsdVolume.Add(oldUsdVolume, volumeInUsd).String()

	// log.Printf("[ updateKLineWithNewData ] debug.. after update, kline: %v\n", kline)
}

func update1DayKline(swap *schema.Swap, mongodb *mongo.Client) {

}

func SaveUpsert1MinKline(kline *schema.KLines1Min, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1MinTableName)

	kline.UpdatedAt = time.Now()

	filter := bson.D{{Key: "symbolDayHour", Value: kline.SymbolDayHour}}
	update := bson.D{{Key: "$set", Value: kline}}
	opt := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(context.Background(), filter, update, opt)
	if err != nil {
		log.Printf("[ SaveUpsert1MinKline ] InsertOne error: %v, kline: %v\n", err, kline)
		return
	} else {
		// log.Printf("[ SaveUpsert1MinKline ] InsertOne success, kline: %v\n", kline)
	}
}

func TEST_KLINE() {
	var pair = schema.KLinePairInfo{
		Symbol:     "pepe/weth",
		BaseToken:  "PEPE",
		QuoteToken: "WETH",
	}
	var k5min = schema.KLines1Min{
		KLinePairInfo: pair,
	}

	// SaveKline(&k5min, chain.GetMongo())

	filter := bson.M{"symbol": "pepe/weth"}
	collection := chain.GetMongo().Database(config.DatabaseName).Collection(config.Kline1MinTableName)

	err := collection.FindOne(context.Background(), filter).Decode(&k5min)

	log.Printf("err: %v, k5min: %v\n", err, k5min)
	log.Println("debug, symbol: ", k5min.Symbol)
}
