package schema

import (
	"sfilter/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
*
每个用户追踪地址发生交易(trader or operator)时, 就生成一条简单记录, 空间换时间
优点: 查询统计都很方便, 把需要排序的数据重复记录即可
缺点: 表行数巨大, 空间也会造成浪费
*/

/*
由于每个用户不允许存储过多记录(会自动删除), 因此即使表很大, 也可接受
1. 假设普通用户1000个用户, 每个用户1000条, 也才100w
2. 设高级用户1000个, 每个用户10000条, 也才1000w
3. 设elite用户100个, 每个用户10w条, 也才1000w
*/

/*
如果删除功能都来不及, 而写入功能已经到达他能够存储条数上限的2倍, 则该用户block掉
*/
type TrackSwap struct {
	UserAddrInfo  `bson:",inline"`
	TrackSwapInfo `bson:",inline"`
	ExtraInfo     `bson:",inline"`
}

// 用户跟踪信息相关
// 相当于重复记录, 避免联表查询
type UserAddrInfo struct {
	Username string `json:"username" bson:"username"`
	Address  string `json:"address" bson:"address"`
	Memo     string `json:"memo" bson:"memo"`
	Priority int    `json:"priority" bson:"priority"`
}

// Swap信息相关
type TrackSwapInfo struct {
	PairAddr string `json:"pairAddr" bson:"pairAddr"` // 交易对地址
	PairName string `json:"pairName" bson:"pairName"`

	MainToken string `json:"mainToken" bson:"mainToken"` // 主token的地址，方便索引
	Direction int    `json:"direction" bson:"direction"` // 买卖方向

	Price             float64 `json:"price" bson:"price"`
	AmountOfMainToken float64 `json:"amountOfMainToken" bson:"amountOfMainToken"`
	VolumeInUsd       float64 ` json:"volumeInUsd" bson:"volumeInUsd"`

	TxHash         string `json:"txhash" bson:"txhash"`
	LogIndexWithTx string `json:"-" bson:"logIndexWithTx"`

	SwapTime time.Time `json:"swapTime" bson:"swapTime"`
}

// 其他信息
type ExtraInfo struct {
	CreatedAt time.Time `json:"-" bson:"createdAt"`
}

var TrackSwapIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "swapTime", Value: -1}},
		Options: options.Index().SetName("swapTime_index").SetExpireAfterSeconds(config.TrackSwapTableSaveTime),
	},
	{
		Keys:    bson.D{{Key: "logIndexWithTx", Value: 1}},
		Options: options.Index().SetName("logIndexWithTx_index").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "username", Value: -1}},
		Options: options.Index().SetName("username_index"),
	},
	{
		Keys:    bson.D{{Key: "address", Value: -1}},
		Options: options.Index().SetName("address_index"),
	},
	{
		Keys:    bson.D{{Key: "memo", Value: -1}},
		Options: options.Index().SetName("memo_index"),
	},
	{
		Keys:    bson.D{{Key: "priority", Value: -1}},
		Options: options.Index().SetName("priority_index"),
	},
	{
		Keys:    bson.D{{Key: "txhash", Value: 1}},
		Options: options.Index().SetName("txhash_index"),
	},
	{
		Keys:    bson.D{{Key: "pairAddr", Value: 1}},
		Options: options.Index().SetName("pairAddr_index"),
	},
	{
		Keys:    bson.D{{Key: "operator", Value: 1}},
		Options: options.Index().SetName("operator_index"),
	},
	{
		Keys:    bson.D{{Key: "trader", Value: 1}},
		Options: options.Index().SetName("trader_index").SetSparse(true),
	},
	{
		Keys:    bson.D{{Key: "mainToken", Value: 1}},
		Options: options.Index().SetName("mainToken_index"),
	},
	{
		Keys:    bson.D{{Key: "volumeInUsd", Value: 1}},
		Options: options.Index().SetName("volumeInUsd_index"),
	},
	{
		Keys:    bson.D{{Key: "amountOfMainToken", Value: 1}},
		Options: options.Index().SetName("amountOfMainToken_index"),
	},
}
