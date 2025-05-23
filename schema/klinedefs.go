package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
	分钟线方案

 	目前全网的 pair 在30w级别, 假设最近2周活跃(否则被删除了)的 pair 为 5w

	方案1: 5min一个点, 一小时12个点(约 12*40 = 480B = 0.5KB); 一天最多288个点;
			 每天一行记录; 设置autodelete, 只存最近14天数据; 也就是最多为 14 * 5w 条记录
			 对于每天来说, 一共24个字段, 每小时一个字段

	方案2: 1min 一个点. 每小时一行记录, 则每天24行, 10天240行; 预计240*5w=1200w条记录
			 对于每小时来说, 最多一共60个点, 也就是 60*40=2.4KB大小

	由于1min可组装5min、15min等所有数据, 总行数也可接受, 因此采纳方案2 - 1min方案
*/

/*
	小时线方案

	假设全网每天活跃的pair在 1w, 每天一行记录; 存180天, 则总行数预计在 1*365 = 365w 级别
	实际上每天活跃pair到不了 1w, 因此ok
*/

/*
	日线方案

	假设全网最终的有效pair(一年无交易的pair清除)在 50w 规模,

	方案1: 每年一行记录; 一行里面存12个月份字段; 暂时定为3年. 最终总行数在 150w 级别

	方案2: 每月一行记录; 假设存3年, 则总行数预计在 50*12*3 = 1800w 级别
			  每行记录存储最多 30或31根柱子, 也就是 30*40 = 1.2KB

	由于1d实际规模不会到50w, 而且千万级别数据可承受, 因此采纳方案2

*/

// Kline 基础结构体，只包含 高开低收
type KLine struct {
	OpenPrice  float64 `bson:"openPrice" json:"openPrice"`
	ClosePrice float64 `bson:"closePrice" json:"closePrice"`
	HighPrice  float64 `bson:"highPrice" json:"highPrice"`
	LowPrice   float64 `bson:"lowPrice" json:"lowPrice"`

	Volume float64 `bson:"volume" json:"volume"`

	// 作用: 由于为了节省表行数, 因此一个字段有多个k线数据
	// 因此增加一个 时间 表示当前KLine表述时间
	// 由于每次udpate的时候, 发现柱子已过期, 会清空当前小时内的所有柱子
	// 所以只要本KLine内有值, 则一定是同一分钟内的update
	UnixTime int64 `bson:"unixTime" json:"unixTime"`

	DeepEyeInfo `bson:",inline"`
}

// deepeye专属额外字段
type DeepEyeInfo struct {
	TxNum       int     `json:"txNum" bson:"txNum"`             // 该周期内的交易数
	VolumeInUsd float64 `bson:"volumeInUsd" json:"volumeInUsd"` // 以usd计价的volume

	PriceInUsd float64 `bson:"priceInUsd" json:"priceInUsd"` // 以usd计价的法币价格
}

type KLinePairInfo struct {
	Pair       string `bson:"pair" json:"pair"`             // 交易对pair地址
	BaseToken  string `bson:"baseToken" json:"baseToken"`   // base 币, 对应为 PEPE
	QuoteToken string `bson:"quoteToken" json:"quoteToken"` // quote计价币, 对应为 WETH
}

type KLineCreatTime struct {
	// 实际最后update时间; 如果回溯, 也为回溯时间
	UpdatedAt time.Time `json:"-" bson:"updatedAt"`

	// 这个Timestamp不表示创建时间, 而是他代表的周期时间
	// 每一次有交易来的时候, 都会更新成其区块时间
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type KLinesForHour [60]KLine // 分钟K线，1小时60根
type KLinesForDay [24]KLine  // 小时K线, 一天24根

type KLinesForMonth [31]KLine // 日K线, 一月最多31根

type KLines1Min struct {
	KLinePairInfo `bson:",inline"` // inline表示内连展开存储

	// Pair_Day组合; 以 Pair 与 day 联合查询到行更新
	PairDayHour string `json:"pairDayHour" bson:"pairDayHour"`

	Kline KLinesForHour `json:"klines" bson:"klines"`

	KLineCreatTime `bson:",inline"`
}

// 小时线
type KLines1Hour struct {
	KLinePairInfo `bson:",inline"`

	PairMonthDay string `json:"pairMonthDay" bson:"pairMonthDay"`

	Kline KLinesForDay `json:"klines" bson:"klines"`

	KLineCreatTime `bson:",inline"`
}

// 日线, 每年一行记录; 暂时不考虑自动删除
type KLines1Day struct {
	KLinePairInfo `bson:",inline"` // inline表示内连展开存储

	// 年份; 通过 pairl与year 查找到具体的行
	PairYearMonth string `json:"pairYearMonth" bson:"pairYearMonth"`

	Kline KLinesForMonth `json:"klines" bson:"klines"`

	KLineCreatTime `bson:",inline"`
}

var Kline1MinIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "timestamp", Value: -1}},
		Options: options.Index().SetName("timestamp_index").SetExpireAfterSeconds(config.Kline1MinTableSaveTime),
	},
	{
		Keys:    bson.D{{Key: "pair", Value: 1}},
		Options: options.Index().SetName("pair_index"),
	},
	{
		Keys:    bson.D{{Key: "pairDayHour", Value: 1}},
		Options: options.Index().SetName("pairDayHour_index").SetUnique(true),
	},
	{
		// base token也就是main token, 需要查询. quoteToken就算了
		Keys:    bson.D{{Key: "baseToken", Value: 1}},
		Options: options.Index().SetName("baseToken_index"),
	},
}

var Kline1HourIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "timestamp", Value: -1}},
		Options: options.Index().SetName("timestamp_index").SetExpireAfterSeconds(config.Kline1HourTableSaveTime),
	},
	{
		Keys:    bson.D{{Key: "pair", Value: 1}},
		Options: options.Index().SetName("pair_index"),
	},
	{
		Keys:    bson.D{{Key: "pairMonthDay", Value: 1}},
		Options: options.Index().SetName("pairMonthDay_index").SetUnique(true),
	},
	{
		// base token也就是main token, 需要查询. quoteToken就算了
		Keys:    bson.D{{Key: "baseToken", Value: 1}},
		Options: options.Index().SetName("baseToken_index"),
	},
}
