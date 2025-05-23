package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LiquidityEvent struct {
	PoolAddress string `json:"poolAddress" bson:"poolAddress"`

	Direction int `json:"direction" bson:"direction"` // 方向, 1为add; 2为remove
	Type      int `json:"type" bson:"type"`

	Amount0 string `json:"amount0" bson:"amount0"`
	Amount1 string `json:"amount1" bson:"amount1"`

	AmountInUsd float64 `json:"amountInUsd" bson:"amountInUsd"`
	PairName    string  `json:"pairName" bson:"pairName"`

	Operator string `json:"operator" bson:"operator"`

	EventBlockNo  uint64    `json:"eventBlockNo" bson:"eventBlockNo"`
	EventTime     time.Time `json:"eventTime" bson:"eventTime"`
	EventTxHash   string    `json:"eventTxHash" bson:"eventTxHash"`
	EventGasPrice string    `json:"eventGasPrice" bson:"eventGasPrice"`

	LogIndexWithTx string `json:"-" bson:"logIndexWithTx"`

	UpdatedAt time.Time `json:"-" bson:"updatedAt"`
	CreatedAt time.Time `json:"-" bson:"createdAt"`
}

var LiquidityEventIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index").SetExpireAfterSeconds(config.LiquidityEventSaveTime),
	},
	{
		Keys:    bson.D{{Key: "poolAddress", Value: 1}},
		Options: options.Index().SetName("poolAddress_index"),
	},
	{
		Keys:    bson.D{{Key: "logIndexWithTx", Value: 1}},
		Options: options.Index().SetName("logIndexWithTx_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "operator", Value: 1}},
		Options: options.Index().SetName("operator_index"),
	},
	{
		Keys:    bson.D{{Key: "eventBlockNo", Value: 1}},
		Options: options.Index().SetName("eventBlockNo_index"),
	},
	{
		Keys:    bson.D{{Key: "amountInUsd", Value: 1}},
		Options: options.Index().SetName("amountInUsd_index"),
	},
	{
		Keys:    bson.D{{Key: "eventTime", Value: 1}},
		Options: options.Index().SetName("eventTime_index"),
	},

	{
		Keys:    bson.D{{Key: "type", Value: 1}},
		Options: options.Index().SetName("type_index"),
	},
}
