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

var kline1hourLock sync.Mutex

func Update1HourKline(swap *schema.Swap, mongodb *mongo.Client) {
	if swap.Price == 0 {
		log.Printf("[ update1MinKline ] wrong price. swap: %v, tx: %v\n", swap, swap.LogIndexWithTx)
		return
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.Kline1HourTableName)

	pair := swap.PairAddr
	quoteToken := swap.Token1
	if swap.MainToken == swap.Token1 {
		quoteToken = swap.Token0
	}

	_time := swap.SwapTime // 以交易(区块)时间为准, 而不是当前时间
	key := fmt.Sprintf("%v_%v_%v", pair, _time.Month(), _time.Day())

	filter := bson.M{"pairMonthDay": key}

	var kline schema.KLines1Hour

	// 直接整体加锁吧, 没啥性能问题
	kline1hourLock.Lock()
	defer kline1hourLock.Unlock()

	err := collection.FindOne(context.Background(), filter).Decode(&kline)
	if err != nil && err != mongo.ErrNoDocuments {
		utils.Warnf("[ update1HourKline ] FindOne error: %v, swap tx: %v\n", err, swap.LogIndexWithTx)
		return
	}

	if kline.PairMonthDay == "" {
		// 说明没有这条k线, 需要新建; 填充基础信息
		kline.Pair = pair
		kline.BaseToken = swap.MainToken
		kline.QuoteToken = quoteToken

		kline.PairMonthDay = key

		kline.Timestamp = _time
	} else {
		// 非新柱子, 则要判断下是否为当前时间柱子, 还是老柱子
		// 跑到这里的时候, 两个时间戳的 Day 一定一样, 那判断他们是否是同一天
		timeDiff := _time.Sub(kline.Timestamp).Abs().Hours()
		if timeDiff > 24 {
			kline.Timestamp = _time
			kline.Kline = schema.KLinesForDay{}
		}
	}

	candisk := &kline.Kline[_time.Hour()]
	updateKLineWithNewData(candisk, swap)

	SaveUpsert1HourKline(&kline, mongodb)
}
