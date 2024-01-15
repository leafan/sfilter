package handler

import (
	"sfilter/config"
	"sfilter/schema"
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
}

// 每隔一定时间执行一遍分析逻辑
func (w *Wiser) Run() {
	schema.InitTables(w.DB)

	interval := time.Duration(w.Config.WiserSearchInterval) * time.Second
	timer := time.NewTicker(interval)
	defer timer.Stop()

	w.WiserSearcher() // 先执行一次

	for range timer.C {
		w.WiserSearcher()
	}
}

func (w *Wiser) WiserSearcher() {
	accounts, err := wiser.GetActiveAccounts(w.Config.AccountActiveSeconds, int64(w.Config.DbBlockReadSize), w.DB)
	if err != nil {
		utils.Errorf("[ WiserSearcher ] GetActiveAccounts error:  %v", err)
		return
	}
	// utils.Infof("[ WiserSearcher ] accounts: %v, err: %v", accounts, err)

	for account := range accounts {
		// 分析出买卖记录并存储
		w.InspectAccountBiDeals(account, accounts[account])

		// 分析账号优秀程度. 与分析买卖独立关系, 甚至可以独立2个进程处理
		w.InspectAccount(account)
	}
}

// 分析某个账号是否是优秀地址
func (w *Wiser) InspectAccount(account string) {
	// utils.Warnf("\n\n[ InspectAccount ] empty function now, todo!!!\n\n")
	// todo...
}

func (w *Wiser) InspectAccountBiDeals(account string, tokens []string) {
	utils.Debugf("[ InspectAccountBiDeals ] account: %v, tokens: %v", account, (tokens))

	for _, token := range tokens {
		atts, err := w.GetAccountTradesForToken(account, token)
		if err != nil {
			continue
		}

		// 将买入卖出记录拆分成一笔一笔的 买卖
		var deals []*schema.BiDeal
		var deal *schema.BiDeal
		var startOver = true // 是否重新计算一笔买卖, 初始化为true

		for _, att := range atts { // atts 按时间从旧到新排列
			if att.Type == schema.WISER_TYPE_SWAP && att.Direction == schema.DIRECTION_BUY_OR_ADD { // 买入且为swap交易
				if startOver {
					deal = &schema.BiDeal{
						Account: account,
						Token:   token,

						BuyTxHash:  att.TxHash,
						BuyBlockNo: att.BlockNo,

						BuyValue:  att.USDValue,
						BuyAmount: att.Amount,
						BuyPrice:  att.PriceInUSD,
					}

					startOver = false // 说明已经有买入动作了, 是一笔新的deal
				} else {
					// 将成本等累加
					deal.BuyValue += att.USDValue
					deal.BuyAmount += att.Amount

					deal.BuyPrice = deal.BuyValue / deal.BuyAmount
				}
			}

			if att.Direction == schema.DIRECTION_SELL_OR_DECREASE { // 只要是sell或者send out
				if startOver || deal == nil {
					// 说明此时统计周期内第一笔就为卖或者已经统计了第一笔卖了, pass

					// utils.Errorf("[ InspectAccountBiDeals ] deal is nil!! please check, account: %v, token: %v, att: %v", account, token, att)
					continue
				}

				// 保存sell信息
				deal.SellTxHash = att.TxHash
				deal.SellBlockNo = att.BlockNo

				deal.SellPrice = att.PriceInUSD

				// 如下两个仅记录当前第一笔sell的时候的金额
				// 但统计的时候会以全部卖出处理
				deal.SellAmount = att.Amount
				deal.SellValue = att.USDValue
				deal.SellType = att.Type

				// 结算deal数据
				deal.EarnChange = (deal.SellPrice - deal.BuyPrice) / deal.BuyPrice
				deal.Earn = deal.EarnChange * deal.BuyValue // earn先以当前价格全部卖出计算

				if deal.SellBlockNo < deal.BuyBlockNo {
					utils.Errorf("[ InspectAccountBiDeals ] wrong block! buyBlock: %v, sellBlock: %v", deal.BuyBlockNo, deal.SellBlockNo)
					continue
				}
				deal.HoldBlocks = deal.SellBlockNo - deal.BuyBlockNo

				// 定义bideal类型
				deal.BiDealType = schema.BI_DEAL_TYPE_CLASSIC

				if deal.HoldBlocks <= uint64(w.Config.ArbitrageBlockInterval) {
					deal.BiDealType = schema.BI_DEAL_TYPE_ARBI
				} else if deal.HoldBlocks <= uint64(w.Config.FrontrunBlockInterval) {
					deal.BiDealType = schema.BI_DEAL_TYPE_FRONTRUN
				}

				deals = append(deals, deal) // 是一笔完整买卖, 存入deals数组

				startOver = true // 标记一笔deal完成, 重新开始新的deal统计
			}
		}

		// 是否要判断下最后一笔的交易状态，如果还持有当前币种未卖出, 则按当前价格计算盈利?
		// 结论: 不判断, 因为这不属于该用户的主动行为

		utils.Infof("[ InspectAccountBiDeals ] len deals: %v", len(deals))
		// wiser.Print(deals)

		for _, deal := range deals {
			wiser.SaveDeal(deal, w.DB)
		}

	}
}

// 获取该用户最近一段时间的swap记录与transfer记录并按要求组装构造格式
func (w *Wiser) GetAccountTradesForToken(account, token string) ([]schema.AccountTokenTrade, error) {
	swapAtts, err1 := wiser.GetAccountSwapsByToken(w.Config.LatestSwapSeconds, w.Config.DbBlockReadSize, account, token, w.DB)

	transferAtts, err2 := wiser.GetAccountTransfersByToken(w.Config.LatestSwapSeconds, w.Config.DbBlockReadSize, account, token, w.DB)

	if err1 != nil || err2 != nil {
		utils.Errorf("[  GetAccountTradesForToken] get account by token error, err1: %v, err2: %v", err1, err2)
		return nil, err1
	}

	// 组装对应的swaps和transfer, 按照时间排序
	// 如果时间一样, 以swap优先排序(因为同一个区块中, swaps比transfer优先级高)
	atts := append(swapAtts, transferAtts...)

	sort.Slice(atts, func(i, j int) bool {
		if atts[i].BlockNo == atts[j].BlockNo {
			// 如果blockNo一样
			if atts[i].BlockNo == atts[j].BlockNo {
				// 如果类型一样, 则保持原来顺序
				if atts[i].Type == atts[j].Type {
					return i < j
				}

				// 如果类型不一样, 则 swap 排到transfer前面, 优先匹配
				return atts[i].Type == schema.WISER_TYPE_SWAP
			}
		}

		return atts[i].BlockNo < atts[j].BlockNo
	})

	// utils.Infof("[ GetAccountTradesForToken ] account: %v, token: %v, atts: %v", account, token, atts)
	return atts, nil
}
