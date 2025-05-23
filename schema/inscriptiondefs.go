package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InscriptionModel struct {
	Data string `json:"data" bson:"data"`

	Inscription           `bson:",inline"`
	TxInfo                `bson:",inline"`
	InscriptionOmitFields `bson:",inline"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// 用于解析铭文结构体
// reference: https://medium.com/coinscription/future-of-erc-20-ethscription-tokens-3bcf234839e3
type Inscription struct {
	Protocol string `json:"p" bson:"p"`
	Tick     string `json:"tick" bson:"tick"`
	Op       string `json:"op" bson:"op"` // deploy or mint

	// deploy params
	Max     string `json:"max" bson:"max"` // max supply
	Limit   string `json:"lim" bson:"lim"` // 每笔mint最大金额
	Decimal string `json:"dec" bson:"dec"` // default 18, 不能超过18

	// mint
	Id     string `json:"id" bson:"id"`   // range: 1-max/lim
	Amount string `json:"amt" bson:"amt"` // amt=lim, 必须相等

}

type InscriptionOmitFields struct {
}

var InscriptionIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index").SetExpireAfterSeconds(config.InscriptionSaveTime),
	},
	{
		Keys:    bson.D{{Key: "p", Value: -1}},
		Options: options.Index().SetName("p_index"),
	},
	{
		Keys:    bson.D{{Key: "tick", Value: -1}},
		Options: options.Index().SetName("tick_index"),
	},
	{
		Keys:    bson.D{{Key: "op", Value: -1}},
		Options: options.Index().SetName("op_index"),
	},
	{
		Keys:    bson.D{{Key: "blockNo", Value: -1}},
		Options: options.Index().SetName("blockNo_index"),
	},
	{
		Keys:    bson.D{{Key: "txHash", Value: -1}},
		Options: options.Index().SetName("txHash_index"),
	},
	{
		Keys:    bson.D{{Key: "Operator", Value: -1}},
		Options: options.Index().SetName("Operator_index"),
	},
}
