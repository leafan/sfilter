package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 全局的一些信息汇总记录与展示
// 存到config 里面, 以key区分
// 该key为: global_info
type GlobalInfo struct {
	ConfigKey string `json:"-" bson:"configKey"`

	TransactionInfo `bson:",inline"`
	OnChainInfo     `bson:",inline"`
	PairCreatedInfo `bson:",inline"`

	UpdatedAt time.Time `json:"-" bson:"updatedAt"`
}

type PairCreatedInfo struct {
	PairCreatedNumIn1h    int     `json:"pairCreatedNumIn1h" bson:"pairCreatedNumIn1h"`
	PairCreatedChangeIn1h float32 `json:"pairCreatedChangeIn1h" bson:"pairCreatedChangeIn1h"`

	PairCreatedNumIn24h    int     `json:"pairCreatedNumIn24h" bson:"pairCreatedNumIn24h"`
	PairCreatedChangeIn24h float32 `json:"pairCreatedChangeIn24h" bson:"pairCreatedChangeIn24h"`
}

type OnChainInfo struct {
	BaseGasPrice         int64   `json:"baseGasPrice" bson:"baseGasPrice"`
	BaseGasPriceChgIn1h  float32 `json:"baseGasPriceChgIn1h" bson:"baseGasPriceChgIn1h"`
	BaseGasPriceChgIn24h float32 `json:"baseGasPriceChgIn24h" bson:"baseGasPriceChgIn24h"`
}

type TransactionInfo struct {
	TxNumIn1h       int     `json:"txNumIn1h" bson:"txNumIn1h"`
	TxNumChangeIn1h float32 `json:"txNumChangeIn1h" bson:"txNumChangeIn1h"`

	TxNumIn24h       int     `json:"txNumIn24h" bson:"txNumIn24h"`
	TxNumChangeIn24h float32 `json:"txNumChangeIn24h" bson:"txNumChangeIn24h"`

	VolumeByUsdIn1h       float64 `json:"volumeByUsdIn1h" bson:"volumeByUsdIn1h"`
	VolumeChangeByUsdIn1h float32 `json:"volumeChangeByUsdIn1h" bson:"volumeChangeByUsdIn1h"`

	VolumeByUsdIn24h       float64 `json:"volumeByUsdIn24h" bson:"volumeByUsdIn24h"`
	VolumeChangeByUsdIn24h float32 `json:"volumeChangeByUsdIn24h" bson:"volumeChangeByUsdIn24h"`
}

// 趋势线, 一分钟插一根, 保存1周即可，用于协助统计 GlobalInfo
type GlobalTrend struct {
	// 以 Day_Hour_Minute作为key即可
	TimelineKey string `json:"timelineKey" bson:"timelineKey"`

	TxNums         int     `json:"txNums" bson:"txNums"`
	PairCreatedNum int     `json:"pairCreatedNum" bson:"pairCreatedNum"`
	BaseGas        int64   `json:"baseGas" bson:"baseGas"` // Int64足够用了
	VolumeByUsd    float64 `json:"volumeByUsd" bson:"volumeByUsd"`

	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // 保存7天差不多了

	CreatedAt time.Time `json:"-" bson:"createdAt"`
}

// 直接创建config表吧，用于自己的查询
var ConfigIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "updatedAt", Value: -1}},
		Options: options.Index().SetName("updatedAt_index"),
	},
	{
		Keys:    bson.D{{Key: "configKey", Value: -1}},
		Options: options.Index().SetName("configKey_index").SetUnique(true),
	},
}

var GlobalTrendIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "timestamp", Value: -1}},
		Options: options.Index().SetName("timestamp_index").SetExpireAfterSeconds(config.GlobalTrendTableSaveTime),
	},
	{
		Keys:    bson.D{{Key: "timelineKey", Value: -1}},
		Options: options.Index().SetName("timelineKey_index"),
	},
}
