package handler

import (
	"fmt"
	"sfilter/schema"
	"sfilter/services/wiser"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
@date: 20240312

买入条件(与):
1. pair创建时间小于1周
2. 第一个小时tx进入前20
3. 5个小时后, 最近1小时tx在前15

卖出条件(或):
1. 五倍卖出
2. 买入后每10分钟取一次tx，tx跌出前20卖出

*/

// hot second pump pair - 第二次重新爆拉热点币
type HSPPair struct {
	Set *Setting
}

func (p *HSPPair) Run() {
	{ // test
		p.TraceBack()
		return
	}

}
func (p *HSPPair) printOneHRank(pairAddr string, hranks []schema.HotPairRank) {
	utils.Infof("\n*** Pair: %v", pairAddr)
	for _, rank := range hranks {
		fmt.Printf("rank: %v, pairAge: %v, createdAt: %v\n", rank.SortRank, rank.PairAge, rank.CreatedAt)
	}

	fmt.Printf("*** End ****\n\n")
}

// 回溯
func (p *HSPPair) TraceBack() {
	hmap := p.GetHotRankMap()

	for pairAddr, hranks := range hmap {
		// p.printOneHRank(pairAddr, hranks)

		p.DoTraceBackOne(pairAddr, hranks)
	}

}

func (p *HSPPair) DoTraceBackOne(pairAddr string, hranks []schema.HotPairRank) {
	if len(hranks) < 2 {
		return // 排名太少不符合需求
	}

	// 第一个小时tx进入前20, 也就是需要判断第一根柱子是否符合需求
	if hranks[0].PairAge > time.Hour && hranks[0].SortRank <= 20 {
		// utils.Warnf("[ DoTraceBackOne ] the first time age too old or rank not good: Age: %v, Rank: %v", hranks[0].PairAge, hranks[0].SortRank)
		return
	}

	// 针对5个小时后面的柱子, 判断是否存在:
	// pair创建时间小于1周 && 最近1小时tx在前15
	var ind int
	var hrank schema.HotPairRank
	hranks = hranks[1:]
	for ind, hrank = range hranks {
		if hrank.PairAge > 6*time.Hour && hrank.PairAge < 7*24*time.Hour {
			// pair时间符合需求, 再判断tx是否符合排名
			if hrank.SortRank < 15 {
				// 找到买点
				// utils.Infof("[ DoTraceBackOne ] pass  buy point.. pair: %v", pairAddr)
				break
			}
		}
	}

	if ind != len(hranks)-1 {
		// 说明找到买点, 开始找卖点
		hranks = hranks[ind+1:]

		// 1.五倍卖出; 2.买入后每小时取一次tx，tx跌出前20卖出
		fmt.Printf("%v, %v, %v\n", hranks[0].PairName, hranks[0].PairAddress, hranks[0].CreatedAt)

	}

}

// 将所有hrank小时排名从db中取出, 以pair为key, createdAt倒排
// 以pair为key, createdAt倒排保存到map数组
func (p *HSPPair) GetHotRankMap() schema.HRankMap {
	filter := bson.M{}

	// 排名前20的即可
	filter["sortRank"] = bson.M{
		"$lte": 20,
	}
	filter["pairLiquidity"] = bson.M{
		"$gte": 1000,
	}

	// 设个大值, 实际达不到
	limit := int64(24 * 100 * 30)
	options := &options.FindOptions{Limit: &limit}

	sort := bson.D{bson.E{Key: "createdAt", Value: -1}}
	options = options.SetSort(sort)

	hmap, err := wiser.GetHotRankToMap(options, &filter, p.Set.DB)
	if err != nil {
		utils.Fatalf("[ GetHotRankMap ] get error: %v", err)
	}

	return hmap
}
