package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Kline 基础结构体，只包含 高开低收
type KLine struct {
	OpenPrice  float64 `bson:"openPrice" json:"openPrice"`
	ClosePrice float64 `bson:"closePrice" json:"closePrice"`
	HighPrice  float64 `bson:"highPrice" json:"highPrice"`
	LowPrice   float64 `bson:"lowPrice" json:"lowPrice"`
	Volume     float64 `bson:"volume" json:"volume"`
}

type KLinePairInfo struct {
	Symbol     string `bson:"symbol" json:"symbol"`         // 交易对, 如 PEPE/WETH, 大写
	BaseToken  string `bson:"baseToken" json:"baseToken"`   // base 币, 对应为 PEPE
	QuoteToken string `bson:"quoteToken" json:"quoteToken"` // quote计价币, 对应为 WETH
}

// 5min一个点, 一小时12个点(约 12*40 = 480B = 0.5KB); 一天最多288个点;
// 每天一行记录; 设置autodelete, 只存最近14天数据; 也就是最多为 14 * tokenNum 条记录
// 对于每天来说, 一共24个字段, 每小时一个字段
// KLine5Min 结构体

type KLines []*KLine

type KLines5Min struct {
	KLinePairInfo `bson:",inline"` // inline表示内连展开存储

	// Symbol_Day组合; 以 symbol 与 day 联合查询到行更新
	SymbolDay string `json:"symbolDay" bson:"symbolDay"`

	// 一小时一个数组
	Hour0 KLines `json:"hour0" bson:"hour0"`
	Hour1 KLines `json:"hour1" bson:"hour1"`
	Hour2 KLines `json:"hour2" bson:"hour2"`
	Hour3 KLines `json:"hour3" bson:"hour3"`
	Hour4 KLines `json:"hour4" bson:"hour4"`
	Hour5 KLines `json:"hour5" bson:"hour5"`
	Hour6 KLines `json:"hour6" bson:"hour6"`
	Hour7 KLines `json:"hour7" bson:"hour7"`
	Hour8 KLines `json:"hour8" bson:"hour8"`
	Hour9 KLines `json:"hour9" bson:"hour9"`

	Hour10 KLines `json:"hour10" bson:"hour10"`
	Hour11 KLines `json:"hour11" bson:"hour11"`
	Hour12 KLines `json:"hour12" bson:"hour12"`
	Hour13 KLines `json:"hour13" bson:"hour13"`
	Hour14 KLines `json:"hour14" bson:"hour14"`
	Hour15 KLines `json:"hour15" bson:"hour15"`
	Hour16 KLines `json:"hour16" bson:"hour16"`
	Hour17 KLines `json:"hour17" bson:"hour17"`
	Hour18 KLines `json:"hour18" bson:"hour18"`
	Hour19 KLines `json:"hour19" bson:"hour19"`

	Hour20 KLines `json:"hour20" bson:"hour20"`
	Hour21 KLines `json:"hour21" bson:"hour21"`
	Hour22 KLines `json:"hour22" bson:"hour22"`
	Hour23 KLines `json:"hour23" bson:"hour23"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`

	// 这个createdAt不表示创建时间, 而是他代表的周期时间
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// 日线, 每年一行记录; 暂时不考虑自动删除
type KLines1Day struct {
	KLinePairInfo `bson:",inline"` // inline表示内连展开存储

	// 年份; 通过 symbol与year 查找到具体的行
	SymbolYear string `json:"symbolYear" bson:"symbolYear"`

	// 每月一个数组, 每个字段最多 31*40 = 1.2KB
	Month1  KLines `json:"month1" bson:"month1"`
	Month2  KLines `json:"month2" bson:"month2"`
	Month3  KLines `json:"month3" bson:"month3"`
	Month4  KLines `json:"month4" bson:"month4"`
	Month5  KLines `json:"month5" bson:"month5"`
	Month6  KLines `json:"month6" bson:"month6"`
	Month7  KLines `json:"month7" bson:"month7"`
	Month8  KLines `json:"month8" bson:"month8"`
	Month9  KLines `json:"month9" bson:"month9"`
	Month10 KLines `json:"month10" bson:"month10"`

	Month11 KLines `json:"month11" bson:"month11"`
	Month12 KLines `json:"month12" bson:"month12"`
}

// 0为read; 1为write
func Get5MinFieldByHour(klines *KLines5Min, hour int) KLines {
	switch hour {
	case 0:
		return klines.Hour0
	case 10:
		return klines.Hour10

	default:

	}

	return klines.Hour23
}

var Kline5MinIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index").SetExpireAfterSeconds(config.Kline5MinTableSaveTime),
	},
	{
		Keys:    bson.D{{Key: "symbol", Value: 1}},
		Options: options.Index().SetName("symbol_index"),
	},
	{
		Keys:    bson.D{{Key: "symbolDay", Value: 1}},
		Options: options.Index().SetName("symbolDay_index").SetUnique(true),
	},
	{
		// base token也就是main token, 需要查询. quoteToken就算了
		Keys:    bson.D{{Key: "baseToken", Value: 1}},
		Options: options.Index().SetName("baseToken_index").SetUnique(true),
	},
}

var Kline1DayIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index").SetExpireAfterSeconds(config.Kline1DayTableSaveTime),
	},
	{
		// symbol 应该是唯一的
		Keys:    bson.D{{Key: "symbol", Value: 1}},
		Options: options.Index().SetName("symbol_index"),
	},
	{
		// symbol 应该是唯一的
		Keys:    bson.D{{Key: "symbolYear", Value: 1}},
		Options: options.Index().SetName("symbolYear_index").SetUnique(true),
	},
	{
		// base token也就是main token, 需要查询. quoteToken就算了
		Keys:    bson.D{{Key: "baseToken", Value: 1}},
		Options: options.Index().SetName("baseToken_index").SetUnique(true),
	},
}
