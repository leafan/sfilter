package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**

几个方案:
	1. 用户通过一个内嵌数组关联所有他跟踪的内容（收藏）
	优点: 行数少, 内容精简
	缺点: 查询及其不方便, 排序困难

	2. 每个用户感兴趣的数据发生时, 就生成一条简单记录, 空间换时间
	优点: 查询统计都很方便, 把需要排序的数据重复记录即可
	缺点: 表行数巨大, 空间也会造成浪费

	3. swap表新建n个index索引字段, 每个用户都生成一个, 如果某条记录属于其感兴趣的, 就加一个字段并加索引
	优点: 省空间, 查询最优最快; 这也是我最先想到的方案
	缺点: 创建用户时麻烦、不优雅

从正常的设计角度来说，应该采用第二种方案, 虽然缺点明显, 但是优雅，扩展性好

研究了下mongo的一些性能分析，如: https://z.itpub.net/article/detail/E6934AD9CFD7A8A6E00CE78CC129D986 64GB内存都支持400亿数据了, 决定就用方案2. 设：每个用户允许存储最近30天数据, 总活跃用户数为1000个。目前每天产生的swap数为 50w 左右
	1. 每个用户如果跟踪地址产生 1000条数据(很多了, 基本看不过来)，那30天共3w条数据, 总记录为3000w，可接受
	2. 每个用户每天产生10000条数据(相当于为了做量化), 那30天共30w条数据, 总记录为3亿，暂不可接受

因此设计为:
	a. 允许跟踪10个地址, 最多允许1000条记录, 及最近30天, $4.99 or 免费?
	b. 允许跟踪100个地址，最多允许1w条记录, 及最近30天  $19.9
	c. 允许跟踪1000个地址, 最多允许10w条记录, 及最近30天; 提供apiKey, 自己实时获取tx; 提供开发指导  $99.9

*/

// 用户对应trackAddress的map表: map[ Trader ] : UserTrackedAddress
// 假设有1000个用户, 平均100个地址, 也就是10w地址, 每个地址 80Byte
// 则一共占用 80*10*10000 = 8MB 而已
type UserTrackAddressMap map[string][]UserTrackedAddress

type UserTrackedAddress struct {
	Username string `json:"username" bson:"username"` // 按username来吧, 可读性高一些, 都是唯一的

	AddressInfo `json:",inline" bson:",inline"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedAt time.Time `json:"-" bson:"createdAt"`
}

type AddressInfo struct {
	Address  string `json:"address" bson:"address"`
	Memo     string `json:"memo" bson:"memo"`
	Priority int    `json:"priority" bson:"priority"`
}

var TrackAddressIndexModel = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "updatedAt", Value: -1}},
		Options: options.Index().SetName("updatedAt_index"),
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
}
