package handler

import (
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/pair"
	"sfilter/services/token"
	"sfilter/services/wiser"
	"sfilter/utils"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

/**
I. 获取活跃地址
	a. 取出最近 x 笔交易，默认7天，目的是为了获取最近交易过的地址
	b. 获取trader即可, 能取到真实trader就取真实的，否则就是 operator
	c. 只要指定时间内发生过swap, 就认为是活跃地址，可以取出来分析
II. 整理地址交易与transfer列表:
	a. 获取该地址最近1个月的所有swap，创建一个该用户交易过的token的map list，key为token
	b. 补充map的value, 每个用户每个token下维护两个list: swap list 与 transfer list，按时间倒序排列
III. 梳理出每一次买卖(参见一次买卖的定义，实现细节逻辑)
IV. 针对每一次买卖做计算、梳理、统计
V. 针对每个用户做计算、梳理、统计
*/

/*
	一次买卖的定义:

	一次买卖是指买入某个token，然后多久之后完全卖出的一个完整动作；对于某个 token, 他可能重复买卖多次(如做波段)
	1. 先找出某个token的第一笔买入动作，作为起始点
	2. 找出买入动作后的第一笔卖出动作或者transfer out 动作
	3. 在这两个动作之间的买入，都作为一次买入，累加成本
	4. 第一笔卖出动作时候，记录此时的价格作为卖出价格计算盈利
	5. 如上1-4步作为一次完整买卖记录下来
	6. 如果到当前时间，还一次未卖过, 该买卖不参与计算
*/

/*
[ bot 识别 ]
arbi: 基本同区块进出, 因此很好判断，对一次买卖做定义即可
frontrun: 基本同区块进出，或者5个block以内的进出，对一次买卖做定义即可
*/

/*
	todo:
 	1. 争议点: 是否要把多笔卖出统计, 而不是第一笔就算卖出
	2. 通缩币是否有问题

*/

type Wiser struct {
	DB     *mongo.Client
	Config *config.WiserConfig

	Tokens schema.TokenMap
	Pairs  schema.PairMap
}

func (w *Wiser) InitMap() {
	var err error
	w.Tokens, err = token.GetTokenMap(w.Config.DbBlockReadSize, w.DB.Database(config.DatabaseName))
	if err != nil {
		utils.Fatalf("[ InitMap ] GetTokenMap failed: %v", err)
	}

	w.Pairs, err = pair.GetPairMap(w.Config.DbBlockReadSize, w.DB.Database(config.DatabaseName))
	if err != nil {
		utils.Fatalf("[ InitMap ] GetPairMap failed: %v", err)
	}

	utils.Infof("[ InitMap ] tokens len: %v, pairs len: %v", len(w.Tokens), len(w.Pairs))
}

// 每隔一定时间执行一遍分析逻辑
func (w *Wiser) Run() {
	schema.InitTables(w.DB)
	w.InitMap()

	interval := time.Duration(w.Config.WiserSearchInterval) * time.Second
	timer := time.NewTicker(interval)
	defer timer.Stop()

	w.WiserSearcher() // 先执行一次

	for range timer.C {
		w.WiserSearcher()
	}
}

func (w *Wiser) WiserSearcher() {
	var accounts []string
	if w.Config.DebugAccount != "" {
		accounts = append(accounts, w.Config.DebugAccount)
	} else {
		var err error
		accounts, err = wiser.GetActiveAccounts(w.Config.AccountActiveSeconds, w.DB)
		if err != nil {
			utils.Fatalf("[ WiserSearcher ] GetActiveAccounts error:  %v", err)
			return
		}
	}
	utils.Infof("[ WiserSearcher ] accounts len: %v", len(accounts))

	for _, account := range accounts {
		// 分析出买卖记录并存储
		w.InspectAccountBiDeals(account)

		// 分析账号优秀程度. 与分析买卖独立关系, 甚至可以独立2个进程处理
		w.InspectAccount(account)
	}
}

// 分析某个账号是否是优秀地址
func (w *Wiser) InspectAccount(account string) {
	utils.Warnf("\n\n[ InspectAccount ] empty function now, todo!!!\n\n")
	// todo...
}

func (w *Wiser) InspectAccountBiDeals(account string) {
	trades, err := w.GetAccountTrades(account)
	if err != nil {
		utils.Warnf("[ InspectAccountBiDeals ] GetAccountTrades err: %v", err)
		return
	}

	for token, atts := range trades {
		tokenObj, ok := w.Tokens[token]
		if !ok {
			utils.Warnf("[ InspectAccountBiDeals ] failed to get token: %v, err: %v", token, err)
			continue
		}

		// 将买入卖出记录拆分成一笔一笔的 买卖
		var deals []*schema.BiDeal
		var deal *schema.BiDeal
		var startOver = true // 是否重新计算一笔买卖, 初始化为true

		for _, att := range atts { // atts 按时间从旧到新排列
			if att.Type == schema.WISER_TYPE_SWAP && att.Direction == schema.DIRECTION_BUY_OR_ADD && att.Amount > 0 { // 买入且为swap交易
				if startOver {
					deal = &schema.BiDeal{
						Account:   account,
						Token:     token,
						TokenName: tokenObj.Name,

						BuyTxHash:  att.TxHash,
						BuyBlockNo: att.BlockNo,
						BuyPair:    att.Pair,

						BuyValue:  att.USDValue,
						BuyAmount: att.Amount,
						BuyPrice:  att.PriceInUSD,
					}

					deal.BuyPairAge = 60 * 60 * 24 * 365 // 初始化为1年
					_pair, ok := w.Pairs[deal.BuyPair]
					if ok {
						bornAt := _pair.FirstAddPoolTime
						if !bornAt.IsZero() { // 存在
							deal.BuyPairAge = int(att.TradeTime.Sub(bornAt).Seconds())
						}

					}

					startOver = false // 说明已经有买入动作了, 是一笔新的deal
				} else {
					// 将成本等累加
					deal.BuyValue += att.USDValue
					deal.BuyAmount += att.Amount

					if deal.BuyAmount > 0 {
						deal.BuyPrice = deal.BuyValue / deal.BuyAmount
					}
				}
			}

			if att.Direction == schema.DIRECTION_SELL_OR_DECREASE { // 只要是sell或者send out
				if deal == nil {
					// 说明周期内还没有买, 是卖的交易
					continue
				}

				if startOver {
					// 说明此时统计周期内第2+笔卖的交易, 累加卖的数据
					// todo..

					continue
				}

				if att.Type == schema.WISER_TYPE_TRANSFER {
					continue // 先不计算, 当前的transfer的金额不准确, 很多数据是回溯的
				}

				if att.PriceInUSD <= 0 {
					utils.Warnf("[ InspectAccountBiDeals ] att.PriceInUSD is 0! token: %v", token)
					// 此时本应该算是一笔完整交易, 但是由于没有价格, 因此不统计与保存
					// 重置状态并开始下一笔deal统计
					startOver = true
					continue
				}

				// 保存sell信息
				deal.SellTxHashWithToken = fmt.Sprintf("%v_%v", att.TxHash, token)
				deal.SellBlockNo = att.BlockNo
				deal.SellPair = att.Pair

				deal.SellPrice = att.PriceInUSD

				// 如下两个仅记录当前第一笔sell的时候的金额
				// 但统计的时候会以全部卖出处理
				deal.SellAmount = att.Amount
				deal.SellValue = att.USDValue
				deal.SellType = att.Type

				// 结算deal数据
				if deal.BuyPrice > 0 {
					deal.EarnChange = (deal.SellPrice - deal.BuyPrice) / deal.BuyPrice
				} else {
					deal.EarnChange = 0 // 0表示异常, 正常情况, 不可能刚好相等
				}

				// 盈亏只算卖出那一笔订单的amount
				deal.Earn = deal.SellAmount * deal.BuyPrice * deal.EarnChange

				if deal.SellBlockNo < deal.BuyBlockNo {
					utils.Errorf("[ InspectAccountBiDeals ] wrong block! buyBlock: %v, sellBlock: %v", deal.BuyBlockNo, deal.SellBlockNo)
					continue
				}

				// 定义bideal类型
				deal.HoldBlocks = deal.SellBlockNo - deal.BuyBlockNo
				deal.BiDealType = w.getDealType(deal.HoldBlocks)

				if w.Config.DebugAccount != "" {
					wiser.PrintDeal(deal)
				}

				deals = append(deals, deal) // 是一笔完整买卖, 存入deals数组

				startOver = true // 标记一笔deal完成, 重新开始新的deal统计
			}
		}

		// 是否要判断下最后一笔的交易状态，如果还持有当前币种未卖出, 则按当前价格计算盈利?
		// 结论: 不判断, 因为这不属于该用户的主动行为

		utils.Infof("[ InspectAccountBiDeals ] token %v has %v deals.", token, len(deals))

		for _, deal := range deals {
			wiser.SaveDeal(deal, w.DB)
		}

	}
}

// 获取该用户最近一段时间的swap记录与transfer记录并按要求组装构造格式
func (w *Wiser) GetAccountTrades(account string) (schema.AccountTrades, error) {
	swapAtts, err1 := wiser.GetAccountSwaps(w.Config.LatestSwapSeconds, w.Config.DbBlockReadSize, account, w.DB)

	transferAtts, err2 := wiser.GetAccountTransfers(w.Config.LatestSwapSeconds, w.Config.DbBlockReadSize, account, w.DB)

	if err1 != nil || err2 != nil {
		utils.Errorf("[  GetAccountTrades] get account by token error, err1: %v, err2: %v", err1, err2)
		return nil, err1
	}

	// 组装对应的swaps和transfer, 按照时间排序
	// 如果时间一样, 以swap优先排序(因为同一个区块中, swaps比transfer优先级高)
	// 合并swap数组, transfer有swap没有的不管
	atts := make(schema.AccountTrades)
	for token, trades := range swapAtts {
		atts[token] = append(atts[token], trades...)

		// 如果有transfer也合并
		transfers, ok := transferAtts[token]
		if ok {
			atts[token] = append(atts[token], transfers...)

			// 排序
			sort.Slice(atts[token], func(i, j int) bool {
				if atts[token][i].BlockNo == atts[token][j].BlockNo {
					// 有些交易是同一个区块里面有买有卖(eg: frontrun), 因此需要处理
					// 如果hash不一样, 按Position排序, 如果hash都一样, 则swap优先
					// 如果一笔交易多笔swap, 会出现transfer排在多笔swap后边的情况, 但对我们统计应该无影响
					if atts[token][i].BlockNo == atts[token][j].BlockNo {
						// 同一笔交易, 防止有些合约不讲武德, transfer与swap事件触发顺序不一致, 因此以swap优先
						if atts[token][i].TxHash == atts[token][j].TxHash {
							// 如果类型也一样, 按position来, 如果类型不一样, 按swap优先来
							if atts[token][i].Type == atts[token][j].Type {
								return atts[token][i].Position < atts[token][j].Position
							} else {
								return atts[token][i].Type == schema.WISER_TYPE_SWAP
							}
						} else {
							// 如果不是同一个区块, 则按position排序即可
							return atts[token][i].Position < atts[token][j].Position
						}
					}
				}

				// 默认按区块高度排序
				return atts[token][i].BlockNo < atts[token][j].BlockNo
			})
		}
	}

	if w.Config.DebugAccount != "" {
		// 调试模式, 打印
		for token, trades := range atts {
			utils.Infof("[ GetAccountTrades ] account: %v, token: %v, atts: %v", account, token, trades)
		}
	}

	return atts, nil
}

// 根据交易间隔区块数获取 deal type
func (w *Wiser) getDealType(blockInterval uint64) int {
	dealType := schema.BI_DEAL_TYPE_CLASSIC

	if blockInterval <= uint64(w.Config.ArbitrageBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_ARBI
	} else if blockInterval <= uint64(w.Config.FrontrunBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_FRONTRUN
	} else if blockInterval <= uint64(w.Config.GambleBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_GAMBLE_TRADE
	} else if blockInterval <= uint64(w.Config.RushBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_RUSH_TRADE
	}

	return dealType
}
