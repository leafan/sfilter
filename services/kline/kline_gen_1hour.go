package kline

import (
	"fmt"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func Get1HourKlineWithFullGenerated(pair string, end time.Time, days int, mongodb *mongo.Database) []schema.KLine {
	var result []schema.KLine

	end = time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), 0, 0, 0, end.Location())

	klines := Get1HourKlineByPairForDays(pair, end, days, mongodb)

	end = end.Add(59 * time.Minute)

	// 先找到第一根不为0的柱子, 并且append, 找到之后, 如果为0值的, 统一赋值为前一个值
	var nonEmpty = false
	for _, klineDay := range klines {
		for _, kline := range klineDay {
			if !nonEmpty {
				if kline.UnixTime != 0 {
					nonEmpty = true
					result = append(result, kline)
				}
			} else {
				last := result[len(result)-1]
				if kline.OpenPrice == 0 {
					kline.OpenPrice = last.ClosePrice
					kline.ClosePrice = last.ClosePrice
					kline.HighPrice = last.ClosePrice
					kline.LowPrice = last.ClosePrice
					kline.PriceInUsd = last.PriceInUsd

					// 加60min...
					kline.UnixTime = last.UnixTime + 60*60
				}

				// 如果当前时间超出end, 则返回, 但不管start
				if kline.UnixTime > end.Unix() {
					goto finish
				}

				result = append(result, kline)
			}
		}
	}

finish:
	// 判断如果柱子不够最少长度, 则拼接生成
	if len(result) < days*24 {
		if len(result) > 0 {
			delta := days*24 - len(result)

			first := result[0]
			var emptyKlines []schema.KLine
			for i := 0; i < delta; i++ {
				new := schema.KLine{
					OpenPrice:  first.OpenPrice,
					ClosePrice: first.OpenPrice,
					HighPrice:  first.OpenPrice,
					LowPrice:   first.OpenPrice,

					// 其他均为0
				}
				new.PriceInUsd = first.PriceInUsd // 价格必须赋值

				emptyKlines = append(emptyKlines, new)
			}

			result = append(emptyKlines, result...)
		}
	} else {
		result = result[len(result)-days*24:]
	}

	return result
}

func Get1HourKlineByPairForDays(pair string, end time.Time, days int, mongodb *mongo.Database) []schema.KLinesForDay {
	start := end.Add(-time.Duration(days) * time.Hour * 24)
	data := get1HourKlineByPair(pair, start, end, mongodb)

	var klines []schema.KLinesForDay
	for i := 0; i < days+1; i++ {
		var kline schema.KLinesForDay

		// 判断当前时间柱子是否有查出数值
		klineTime := start.Add(time.Duration(i) * time.Hour * 24)
		pairMonthDay := fmt.Sprintf("%v_%v_%v", pair, klineTime.Month(), klineTime.Day())

		for _, v := range data {
			// 如果当前小时有合法柱子, 拷贝, 否则全为0
			if v.PairMonthDay == pairMonthDay {

				utils.DeepCopy(&v.Kline, &kline)
				break
			}
		}

		klines = append(klines, kline)
	}

	return klines
}
