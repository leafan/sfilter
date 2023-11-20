package handler

import (
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/kline"
	"sfilter/services/token"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func HandleTokenInfo(block *schema.Block, mongodb *mongo.Client, swaps []*schema.Swap) {
	// 先取出本次需要更新的token信息
	tokens := make(map[string]int)

	for _, swap := range swaps {
		tokens[swap.MainToken]++
	}
	// log.Printf("[ HandleTokenInfo ] tokens: %v\n\n", tokens)

	for key := range tokens {
		updateTokenInfo(key, mongodb)
	}

}

func updateTokenInfo(tokenAddr string, mongodb *mongo.Client) {
	log.Printf("[ updateTokenInfo ] token: %v\n\n", tokenAddr)

	//  baseToken如果是u币, 不计算
	if utils.CheckExistString(tokenAddr, config.QuoteUsdCoinList) {
		return
	}

	_token, err := chain.GetTokenInfo(tokenAddr)
	if err != nil {
		log.Printf("[ updateTokenInfo ] GetTokenInfo wrong, return.. err: %v\n\n", err)
		return
	}
	log.Printf("[ updateTokenInfo ] token: %v\n", _token)

	// 为避免mongo反复读取, 使用kline相关信息, 然后进行计算更新
	// 只需要2次get操作(分钟线与小时线), 就可以全部计算完毕

	// 取最近180分钟数据, 可以计算出最少2h(某时刚过1min时需要多1h数据)内的数据变化(Change)
	// 可能取出多个交易对的数据, tx和volume累加, price以U计价取
	now := time.Now()
	klines1Min := kline.Get1MinKlineByPairForHours(_token.Address, now, 2, mongodb)
	// log.Printf("[ updateTokenInfo ] token: %v, klines1Min: %v\n", _token.Address, klines1Min)

	updateTokenTx(&klines1Min, _token, now)
	updateTokenPrice(_token)
	updateTokenVolume(_token)

	// update to db
	token.UpdateTokenInfo(_token, mongodb)
}

func updateTokenTx(klines1Min *[]schema.KLinesForHour, _token *schema.Token, now time.Time) {
	// klines1Min 为最近3小时内的柱子,且按时间逆序
	// var curHourKline []schema.KLine  // 最近1小时的kline
	// var lastHourKline []schema.KLine // 上一个小时的kline

	// curMin = now.Minute()
	// if klines1Min[0]

	// 首先如果没有取到数据

}

func updateTokenPrice(_token *schema.Token) {

}

func updateTokenVolume(_token *schema.Token) {

}
