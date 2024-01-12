package kline

import (
	"context"
	"fmt"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var kline1minLock sync.Mutex

func Update1MinKline(swap *schema.Swap, mongodb *mongo.Client) {
	// 防止同时写入的时候, 查询时覆盖
	// 简单实现, 实际上应该是给 行加读写锁, 这里直接给整个加锁了..

	if swap.Price == 0 {
		log.Printf("[ update1MinKline ] wrong price. swap: %v, tx: %v\n", swap, swap.LogIndexWithTx)
		return
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1MinTableName)

	pair := swap.PairAddr
	quoteToken := swap.Token1
	if swap.MainToken == swap.Token1 {
		quoteToken = swap.Token0
	}

	_time := swap.SwapTime // 以交易(区块)时间为准, 而不是当前时间
	key := fmt.Sprintf("%v_%v_%v", pair, _time.Day(), _time.Hour())

	// log.Printf("[ update1MinKline ] come here key: %v, minute: %v, swap: %v\n", key, _time.Minute(), swap)

	filter := bson.M{"pairDayHour": key}

	var kline schema.KLines1Min

	// 直接整体加锁吧, 没啥性能问题
	kline1minLock.Lock()
	defer kline1minLock.Unlock()

	err := collection.FindOne(context.Background(), filter).Decode(&kline)
	if err != nil && err != mongo.ErrNoDocuments {
		utils.Warnf("[ update1MinKline ] FindOne error: %v, swap tx: %v\n", err, swap.LogIndexWithTx)
		return
	}

	if kline.PairDayHour == "" {
		// 说明没有这条k线, 需要新建; 填充基础信息
		kline.Pair = pair
		kline.BaseToken = swap.MainToken
		kline.QuoteToken = quoteToken

		kline.PairDayHour = key

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
	curPrice := swap.Price

	kline.ClosePrice = curPrice // 不管新老柱子, 先更新close
	kline.PriceInUsd = swap.PriceInUsd

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

	kline.UnixTime = swap.SwapTime.Unix()

	// volume啥都不用考虑, 直接加
	volume := swap.AmountOfMainToken
	kline.Volume += volume

	// 更新 deepeye info
	kline.TxNum++
	kline.VolumeInUsd += swap.VolumeInUsd

	// log.Printf("[ updateKLineWithNewData ] debug.. after update, kline: %v\n", kline)
}
