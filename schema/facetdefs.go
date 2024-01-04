package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
 * 针对前置data为: data:application/vnd.facet.tx+json;rule=esip6,
 * 做facet(默认这个前缀为facet定义)业务解析
 */

type FacetModel struct {
	FacetProjectInfo `bson:",inline"` // 项目信息, 根据 to 目标地址确认
	TxInfo           `bson:",inline"` // 交易相关信息
	UserLogic        `bson:",inline"` // json data里面的具体信息解析

	FacetOmitFields `bson:",inline"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type FacetProjectInfo struct {
	// 项目地址, 如: 0x00000000000000000000000000000000000FacE7
	ProjectAddress string `json:"projectAddress" bson:"projectAddress"`
	// 项目名称, 如 Facet. 方便查看
	ProjectName string `json:"projectName" bson:"projectName"`
}

type TxInfo struct {
	BlockNo uint64 `json:"blockNo" bson:"blockNo"` // 区块号
	TxHash  string `json:"txHash" bson:"txHash"`   // 交易哈希

	Operator string `json:"operator" bson:"operator"` // msg.sender
	GasPrice string `json:"gasPrice" bson:"gasPrice"` // 使用的gas price
}

type UserLogic struct {
	Data string `json:"data" bson:"data"` // 原始的json data, 保存下来

	Op string `json:"op" bson:"op"` // 操作类型, 一般为call等

	// 如下为其data解析
	Function string `json:"function" bson:"function"`
	To       string `json:"to" bson:"to"` // 如 facetswap地址 0xf29e6e319ac4ce8c100cfc02b1702eb3d275029e

	Args `bson:",inline"`
}

// 里面的args解析, 能解析就解析
type Args struct {
	ArgsTo     string  `json:"argsTo" bson:"argsTo"`         // 如sell token 的收币地址
	ArgsAmount float64 `json:"argsAmount" bson:"argsAmount"` // amount 或 amountIn
}

// 内部使用, 不保存到db
type FacetOmitFields struct {
}

// 用于解析json结构体
type FacetJson struct {
	Op   string        `json:"op"`
	Data FacetDataJson `json:"data"`
}
type FacetDataJson struct {
	To       string                 `json:"to"`
	Function string                 `json:"function"`
	Args     map[string]interface{} `json:"args"` // 有一些如 approve 目前也不识别
}

var FacetIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("createdAt_index").SetExpireAfterSeconds(config.FacetSaveTime),
	},
	{
		Keys:    bson.D{{Key: "projectAddress", Value: -1}},
		Options: options.Index().SetName("projectAddress_index"),
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
	{
		Keys:    bson.D{{Key: "function", Value: -1}},
		Options: options.Index().SetName("function_index"),
	},
	{
		Keys:    bson.D{{Key: "op", Value: -1}},
		Options: options.Index().SetName("op_index"),
	},
	{
		Keys:    bson.D{{Key: "to", Value: -1}},
		Options: options.Index().SetName("to_index"),
	},
	{
		Keys:    bson.D{{Key: "argsTo", Value: -1}},
		Options: options.Index().SetName("argsTo_index"),
	},
	{
		Keys:    bson.D{{Key: "ArgsAmount", Value: -1}},
		Options: options.Index().SetName("ArgsAmount_index"),
	},
}
