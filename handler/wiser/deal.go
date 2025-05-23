package handler

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

import (
	"errors"
	"fmt"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/pair"
	"sfilter/services/token"
	"sfilter/services/wiser"
	"sfilter/utils"
	"sort"
	"time"
)

func (w *Wiser) InspectBiDeals(account string) int {
	utils.Infof("[ InspectBiDeals ] account: %v", account)

	dealCount := 0

	trades, err := w.GetAccountTrades(account)
	if err != nil {
		utils.Errorf("[ InspectBiDeals ] GetAccountTrades err: %v", err)
		return dealCount
	}

	for token, atts := range trades {
		tokenObj, ok := w.set.Tokens[token]
		if !ok {
			tokenObj, err = chain.GetTokenInfo(token, w.set.DB)
			if err != nil {
				utils.Errorf("[ InspectBiDeals ] failed to get token: %v, err: %v", token, err)
				continue
			}
		}

		deals := w.getDealsFromAtts(atts, account, tokenObj)

		utils.Debugf("[ InspectBiDeals ] token %v has %v deals.", token, len(deals))
		dealCount += len(deals)

		for _, deal := range deals {
			wiser.SaveDeal(deal, w.set.DB)
		}

	}

	return dealCount
}

func (w *Wiser) getDealsFromAtts(atts []schema.AccountTokenTrade, account string, tokenObj *schema.Token) []*schema.BiDeal {
	var deals []*schema.BiDeal
	var deal *schema.BiDeal
	var startOver = true // 是否重新计算一笔买卖, 初始化为true

	for _, att := range atts { // atts 按时间从旧到新排列
		if att.Type == schema.TRADE_TYPE_SWAP && att.Direction == schema.DIRECTION_BUY_OR_ADD && att.Amount > 0 { // 买入且为swap交易
			if startOver {
				deal = &schema.BiDeal{
					Account:   account,
					Token:     tokenObj.Address,
					TokenName: tokenObj.Name,

					BuyTxHash:   att.TxHash,
					BuyBlockNo:  att.BlockNo,
					BuyPair:     att.Pair,
					BuyPairType: att.PairType,

					BuyValue:  att.USDValue,
					BuyAmount: att.Amount,
					BuyPrice:  att.PriceInUSD,
				}

				deal.BuyPairAge = 60 * 60 * 24 * 181 // 初始化为1年
				_pair, ok := w.set.Pairs[deal.BuyPair]
				if !ok {
					var err error
					_pair, err = pair.GetPairInfoForRead(deal.BuyPair)
					if err == nil {
						ok = true
					} else {
						utils.Errorf("[ getDealsFromAtts ] GetPair %v failed: %v", deal.BuyPair, err)
					}
				}
				if ok {
					bornAt := _pair.FirstAddPoolTime
					if !bornAt.IsZero() { // 存在
						deal.BuyPairAge = int(att.TradeTime.Sub(bornAt).Seconds())
					}

					deal.BuyPairHackType = _pair.MainTokenHackType
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
			if deal == nil { // 说明周期内还没有买, 是卖的交易
				continue
			}
			if startOver { // 说明此时统计周期内第2+笔卖的交易, 累加卖的数据
				// todo
				continue
			}

			if att.Type == schema.TRADE_TYPE_TRANSFER {
				continue // 先不计算, 当前的transfer的金额不准确, 很多数据是回溯的
			}
			if att.PriceInUSD <= 0 {
				utils.Warnf("[ InspectBiDeals ] att.PriceInUSD is 0! token: %v", tokenObj.Address)
				// 此时本应该算是一笔完整交易, 但是由于没有价格, 因此不统计与保存
				// 重置状态并开始下一笔deal统计
				startOver = true
				continue
			}

			// 保存sell信息
			deal.SellTxHashWithToken = fmt.Sprintf("%v_%v", att.TxHash, tokenObj.Address)
			deal.SellBlockNo = att.BlockNo
			deal.SellTime = att.TradeTime
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
				utils.Errorf("[ InspectBiDeals ] wrong block! buyBlock: %v, sellBlock: %v", deal.BuyBlockNo, deal.SellBlockNo)
				continue
			}

			// 定义bideal类型
			deal.HoldBlocks = deal.SellBlockNo - deal.BuyBlockNo
			deal.BiDealType = w.getDealType(deal.HoldBlocks)
			deal.BuyType = w.getBuyType(deal.BuyPairAge)

			// 统计截止至今的亏损率(假设未卖)
			if deal.BuyPrice > 0 {
				uptoTodayUsdValue, err := w.getDealSellValueWithAllBuyAmount(deal)
				if err == nil {
					deal.UptoTodayYield = (uptoTodayUsdValue - deal.BuyValue) / deal.BuyValue
				}
			}

			if w.set.Config.DebugMode {
				wiser.PrintDeal(deal)
			}

			deals = append(deals, deal) // 是一笔完整买卖, 存入deals数组

			startOver = true // 标记一笔deal完成, 重新开始新的deal统计
		}
	}

	// 是否要判断下最后一笔的交易状态，如果还持有当前币种未卖出, 则确认是否已跌到接近归零
	// 如果是, 则计算当前亏钱状态; 否则不记录, 作为无效deal
	if !startOver {
		// 说明此时还持有token, 计算当前盈利状态以确认是否作为有效deal记录
		validDeal := w.updateDealWithLatestData(deal, tokenObj)
		if validDeal {
			deals = append(deals, deal)
		}
	}

	return deals
}

// 确认当前deal持有从未卖出的token, 根据当前价格计算盈利
func (w *Wiser) updateDealWithLatestData(deal *schema.BiDeal, tokenObj *schema.Token) bool {
	// deal无有效买入
	if deal == nil || deal.BuyValue <= 0 || deal.BuyAmount <= 0 || deal.Account == "" {
		return false
	}

	if deal.SellTxHashWithToken != "" {
		// deal 已经收尾, 已经有了卖的交易
		return false
	}

	// utils.Debugf("[ updateDealWithLatestData ] found one deal to be updated. deal: %v", deal)

	// 继续往后走, 说明该用户对该token有买入动作但无有效的卖出动作(transfer无效)
	deal.SellType = schema.TRADE_TYPE_LIQUIDATION
	// 新建一个唯一键值, 用BuyTx+Token+Account代替
	deal.SellTxHashWithToken = fmt.Sprintf("%v_%v_%v", deal.BuyTxHash, deal.Token, deal.Account)

	// 将当前的amount卖到当前的pair里面, 作为earn数据并统计更新
	// 可能出现实际情况错误, 比如transfer出去或者其他没被统计到的情况, 但先忽略
	// 这里也不获取余额, 直接模拟将所有buy的卖出去
	sellUsdValue, err := w.getDealSellValueWithAllBuyAmount(deal)
	if err != nil || sellUsdValue < 0 {
		utils.Warnf("[ updateDealWithLatestData ] get deal amout value failed. err: %v, sellUsdValue: %v", err, sellUsdValue)
		return false
	}

	// 此时需要判断其sell value是否已亏损过多或盈利很高, 如果不是, 也不结算
	if sellUsdValue > deal.BuyValue*w.set.Config.DealDefiniteLoss &&
		sellUsdValue < deal.BuyValue*w.set.Config.DealDefiniteWin {
		utils.Debugf("[ updateDealWithLatestData ] sellUsdValue not win or loss too much. current value: %v, BuyValue: %v", sellUsdValue, deal.BuyValue)
		return false
	}

	deal.SellAmount = deal.BuyAmount
	deal.SellValue = sellUsdValue
	deal.SellPrice = sellUsdValue / deal.SellAmount

	deal.Earn = deal.SellValue - deal.BuyValue
	deal.EarnChange = deal.Earn / deal.BuyValue

	// 定义bideal类型
	deal.SellBlockNo, _ = chain.GetCurrentBlockNumber()
	deal.SellTime = time.Now()

	deal.HoldBlocks = deal.SellBlockNo - deal.BuyBlockNo
	deal.BiDealType = w.getDealType(deal.HoldBlocks)

	deal.BuyType = w.getBuyType(deal.BuyPairAge)

	deal.UptoTodayYield = (sellUsdValue - deal.BuyValue) / deal.BuyValue

	return true
}

func (w *Wiser) getDealSellValueWithAllBuyAmount(deal *schema.BiDeal) (float64, error) {
	pairObj, ok := w.set.Pairs[deal.BuyPair]
	if !ok {
		var err error
		pairObj, err = pair.GetPairInfoForRead(deal.BuyPair)
		if err != nil {
			utils.Errorf("[ updateDealWithLatestData ] find pair failed. pair: %v, err: %v", deal.BuyPair, err)
			return 0, err
		}
	}

	// 买入金额已经转换成为了float, 需要先转成 big.Float，再转成 int.string
	var tokenOut string
	var decimalIn uint8
	var decimalOut uint8
	var tokenExponent *big.Int

	if deal.Token == pairObj.Token0 {
		decimalIn = pairObj.Decimal0
		decimalOut = pairObj.Decimal1
		tokenOut = pairObj.Token1
	} else {
		decimalIn = pairObj.Decimal1
		decimalOut = pairObj.Decimal0
		tokenOut = pairObj.Token0
	}
	tokenExponent = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimalIn)), nil)

	buyAmountF := big.NewFloat(deal.BuyAmount)
	buyAmountF = buyAmountF.Mul(buyAmountF, new(big.Float).SetInt(tokenExponent))
	buyAmountBInt, _ := buyAmountF.Int(nil)

	var amountOut *big.Int
	var errRet error

	if deal.BuyPairType == schema.SWAP_EVENT_UNISWAPV2_LIKE {
		amountOut, errRet = chain.GetUniV2SwapAmountOut(deal.BuyPair, pairObj.Token0, deal.Token, buyAmountBInt)
		if errRet != nil {
			utils.Warnf("[ getDealAmountValueWithAmountIn ] GetUniV2SwapAmountOut failed: %v, pair: %v", errRet, deal.BuyPair)
		}
	} else if deal.BuyPairType == schema.SWAP_EVENT_UNISWAPV3_LIKE {
		var feeBig *big.Int

		fee := pairObj.PairFee
		if fee == 0 {
			feeBig, errRet = chain.GetUniV3PairFee(pairObj.Address)
			if errRet != nil {
				utils.Warnf("[ getDealAmountValueWithAmountIn ] GetUniV3PairFee failed. pair: %v, err: %v", pairObj.Address, errRet)
				goto finish
			}

			fee = feeBig.Int64()
		}

		amountOut, errRet = chain.GetUniV3SwapAmountOut(deal.Token, tokenOut, big.NewInt(fee), buyAmountBInt)
		if errRet != nil {
			utils.Warnf("[ getDealAmountValueWithAmountIn ] GetUniV3SwapAmountOut failed: %v. pair: %v, in: %v, out: %v, fee: %v, amountIn: %v, amountOut: %v", errRet, deal.BuyPair, deal.Token, tokenOut, fee, buyAmountBInt, amountOut)
		}
	} else {
		errRet = errors.New("wrong pair type")
		utils.Errorf("[ getDealAmountValueWithAmountIn ] unknown pair type: %v", deal.BuyPairType)
	}

finish:
	var sellUsdValue = -1.0 // 默认值为负值

	if errRet != nil {
		// 存在错误, 判断下余额, 如果余额比买入时的value还小一定比例, 则作为其买入失败
		liquidityVolume := pairObj.LiquidityInUsd
		if liquidityVolume < deal.BuyValue*w.set.Config.DealDefiniteLoss {
			utils.Infof("[ getDealAmountValueWithAmountIn ] liquidityVolume too less: %v, buyValue: %v", liquidityVolume, deal.BuyValue)
			sellUsdValue = liquidityVolume
			errRet = nil
		}
	} else {
		// 还需要将 tokenOut 的amountOut转成法币
		var ethPrice float64
		ethPrice, errRet = chain.GetBasicCoinPrice(nil, nil, config.BlockChain)
		if errRet == nil {
			sellUsdValue = utils.CalculateVolumeInUsd(tokenOut, new(big.Float).SetInt(amountOut), decimalOut, ethPrice)
		}
	}

	// 如果还为0, 则有可能pair双方都是非价值币, 直接从pair中取当前usd 价格计算
	if sellUsdValue <= 0 && amountOut != nil {
		_token, err := token.GetTokenInfo(tokenOut, w.set.DB)
		if err == nil {
			amountOutF := new(big.Float).SetInt(amountOut)
			tokenExponent := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimalOut)), nil)
			amountOutF = amountOutF.Quo(amountOutF, new(big.Float).SetInt(tokenExponent))

			amount, _ := amountOutF.Float64()

			sellUsdValue = amount * _token.PriceInUsd
		}
	}

	return sellUsdValue, errRet
}

// 获取该用户最近一段时间的swap记录与transfer记录并按要求组装构造格式
func (w *Wiser) GetAccountTrades(account string) (schema.AccountTrades, error) {
	swapAtts, err1 := wiser.GetAccountSwaps(w.set.Config.LatestSwapSeconds, w.set.Config.DbBlockReadSize, account, w.set.DB)

	transferAtts, err2 := wiser.GetAccountTransfers(w.set.Config.LatestSwapSeconds, w.set.Config.DbBlockReadSize, account, w.set.DB)

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
								return atts[token][i].Type == schema.TRADE_TYPE_SWAP
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

	// if w.set.Config.DebugMode {
	// 	// 调试模式, 打印
	// 	for token, trades := range atts {
	// 		utils.Infof("[ GetAccountTrades ] account: %v, token: %v, atts: %v", account, token, trades)
	// 	}
	// }

	return atts, nil
}

// 根据交易间隔区块数获取 deal type
func (w *Wiser) getDealType(blockInterval uint64) int {
	dealType := schema.BI_DEAL_TYPE_TREND

	if blockInterval <= uint64(w.set.Config.ArbitrageBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_ARBI
	} else if blockInterval <= uint64(w.set.Config.FrontrunBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_FRONTRUN
	} else if blockInterval <= uint64(w.set.Config.GambleBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_GAMBLE_TRADE
	} else if blockInterval <= uint64(w.set.Config.RushBlockInterval) {
		dealType = schema.BI_DEAL_TYPE_RUSH_TRADE
	}

	return dealType
}

func (w *Wiser) getBuyType(seconds int) int {
	buyType := schema.BI_DEAL_BUY_TYPE_TREND

	if seconds <= w.set.Config.DealBuyTypeMev {
		buyType = schema.BI_DEAL_BUY_TYPE_MEV
	} else if seconds < w.set.Config.DealBuyTypeFresh {
		buyType = schema.BI_DEAL_BUY_TYPE_FRESH
	} else if seconds < w.set.Config.DealBuyTypeSubNew {
		buyType = schema.BI_DEAL_BUY_TYPE_SUBNEW
	}

	return buyType
}
