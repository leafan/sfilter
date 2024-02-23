package kline

import (
	"fmt"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// 生成1min k线, 合并成一个大数组, 且将0值更新成前值(无交易时等于前值)
// 结束时间为end, 但起始时间可能比end-hours早, 最多不超过1小时
func Get1MinKlineWithFullGenerated(pair string, end time.Time, hours int, mongodb *mongo.Database) []schema.KLine {
	var result []schema.KLine

	// 将end更新到当前分钟的0值
	end = time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), end.Minute(), 0, 0, end.Location())

	klines := get1MinKlineByPairForHours(pair, end, hours, mongodb)

	end = end.Add(59 * time.Second) // 再加1min, 取到当前柱子

	// 先找到第一根不为0的柱子, 并且append, 找到之后, 如果为0值的, 统一赋值为前一个值
	var nonEmpty = false
	for _, klineHour := range klines {
		for _, kline := range klineHour {
			// 还没找到不为0的柱子
			if !nonEmpty {
				if kline.UnixTime != 0 {
					nonEmpty = true
					result = append(result, kline)
				}
			} else {
				// 将当前kline值的open等价格更新然后append
				// 跑到这里来的时候, result一定不为空
				last := result[len(result)-1]
				if kline.OpenPrice == 0 {
					kline.OpenPrice = last.ClosePrice
					kline.ClosePrice = last.ClosePrice
					kline.HighPrice = last.ClosePrice
					kline.LowPrice = last.ClosePrice
					kline.PriceInUsd = last.PriceInUsd

					// 加1分钟
					kline.UnixTime = last.UnixTime + 60
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
	if len(result) < hours*60 {
		if len(result) > 0 {
			delta := hours*60 - len(result)

			first := result[0]
			var emptyKlines []schema.KLine
			for i := 0; i < delta; i++ {
				new := schema.KLine{
					OpenPrice:  first.OpenPrice, // 下一个的open为上一个的close
					ClosePrice: first.OpenPrice,
					HighPrice:  first.OpenPrice,
					LowPrice:   first.OpenPrice,
				}
				new.PriceInUsd = first.PriceInUsd

				emptyKlines = append(emptyKlines, new)
			}

			result = append(emptyKlines, result...)
		}
	} else {
		result = result[len(result)-hours*60:]
	}

	return result
}

// 获取k线, 且倒推几小时且返回值必须生成几根柱子
func get1MinKlineByPairForHours(pair string, end time.Time, hours int, mongodb *mongo.Database) []schema.KLinesForHour {
	start := end.Add(-time.Duration(hours) * time.Hour)
	data := get1MinKlineByPair(pair, start, end, mongodb)

	// 需要生成(hours+1)个柱子, 按时间正序
	// 如 1:10~3:10, 则有1,2,3点钟三根柱子, 即使是 1:00~3:00, 亦是如此
	var klines []schema.KLinesForHour
	for i := 0; i < hours+1; i++ {
		var kline schema.KLinesForHour

		// 判断当前时间柱子是否有查出数值
		klineTime := start.Add(time.Duration(i) * time.Hour)
		pairDayHour := fmt.Sprintf("%v_%v_%v", pair, klineTime.Day(), klineTime.Hour())

		for _, v := range data {
			// 如果当前小时有合法柱子, 拷贝, 否则全为0
			if v.PairDayHour == pairDayHour {
				// log.Printf("[ get1MinKlineByPairForHours ] find one pairDayHour: %v, v: %v\n", pairDayHour, v)

				utils.DeepCopy(&v.Kline, &kline)
				break
			}
		}

		klines = append(klines, kline)
	}

	return klines
}
