package handler

import (
	"fmt"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/pair"
	service_swap "sfilter/services/swap"
	"sfilter/services/token"
	"sfilter/utils"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleSwapAndKline(block *schema.Block, mongodb *mongo.Client) []*schema.Swap {
	var swaps []*schema.Swap

	for _, tx := range block.Transactions {
		if len(tx.Receipt.Logs) > 0 {
			for _, _log := range tx.Receipt.Logs {
				if len(_log.Topics) > 0 {
					_type := checkSwapEvent(_log.Topics)

					// 发现有swap交易
					if _type > 0 {
						// log.Printf("[ Swap_handle ] swap tx now. type: %v, tx: %v\n\n", _type, tx.OriginTx.Hash())

						swap := newSwapStruct(block, _log, tx)
						if swap == nil {
							// new有错误, continue掉
							continue
						}
						swap.SwapType = _type

						if _type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
							updateUniV2Swap(swap, _log, mongodb)
						} else if _type == schema.SWAP_EVENT_UNISWAPV3_LIKE {
							updateUniV3Swap(swap, _log, mongodb)
						}

						updateUsdInfo(swap, mongodb)

						UpdateKlineBySwap(swap, mongodb)

						updateBlockInfo(block, swap)

						swaps = append(swaps, swap)
					}
				}

			}
		}
	}

	return swaps
}

func updateBlockInfo(blk *schema.Block, swap *schema.Swap) {
	blk.TxNums++
	blk.VolumeByUsd += swap.VolumeInUsd
}

/*
@update 20240125:  纯粹通过transfer的流转来判断交易受益方(trader)

针对每一笔tx, 将所有的transfer抽离出来, 组装成如下数据结构, 并将 pair/router 地址打标记(del)或删除:
Token_X1[from] = slice[]... Token_X1[to] = slice[]...
Token_X2[from] = slice[]...  Token_X2[to] = slice[]...
...

I: 然后将 pair/router的 from or to 地址删除(表示中间流转)
III: 把通缩币的常见to地址收集(有困难, 需要其他手段并进), 0地址等处理
II. 如果最终剩余下来的token中 from 和 to 相等, 表示为arbi, trader也为null

 1. 对于普通的单路径交易, 会涉及到2个token的transfer, 如 weth-> usdt
    Weth[ from ] = [ Address_Sell ]; Weth[ to ] = [ Uni_Pair(del) ]
    Usdt[ from ] = [ Uni_Pair(del) ]; Usdt[ to ] = [ Address_Trader ]
    经过去除后, 可以非常明显的找到 trader为 Address_Sell(该pair中为sell, 所以找from)

 2. 对于经过router的单路径交易, 要把deposit的eth作为一笔transfer加进来, from为触发地址, to为router
    Weth[ from ] = [ Operator, Router(del) ];	Weth[ to ] = [ Router(del), Uni_Pair(del) ]
    Token[ from ] = [ Uni_Pair(del) ];	Token[ to ] = [ Token_to ]
    此时非常明显的 weth[from] = operator; token[to] = Token_to 了

    -- 经测试, router的eth也是往eth走, 因此可以忽略, 因此此处不需要实现 deposit eth 的transfer

 3. 对于多路径交易, 比如2个path, 举例 Usdt->Weth->Mubi (https://etherscan.io/tx/0xe6a4bad02035624cac83f0c1698242405c1dc5e49d7a70d4c0f879f03b1daa2a)
    Usdt[ from ] = [ Operator ];	Usdt[ to ] = [ USDT/ETH_Pair(del) ]
    Weth[ from ] = [ USDT/ETH_Pair(del), Router(del) ];	Weth[ to ] = [ Router(del), Mubi_Pair(del) ]
    Mubi[ from ] = [ Mubi_Pair(del) ];	Mubi[ to ] = Trader

    从这里能看到, 经删除后, 只剩下:		Usdt[from], 以及 Mubi[to], 	但是这里有2笔swap:
    a. usdt->eth(买入eth): trader = null. 因为需要找eth的to, 发现没有. 本质确实也是, 他并没有买入eth
    如果这里是 pepe->eth(卖出pepe): trader = Operator. 也是正确的, 他确实卖出了
    b. eth->mubi(买入mubi): trader = mubi[to], 正确

 4. 对于arbi, 如: Weth->Edoge->Weth(https://etherscan.io/tx/0xb4b8c81d0a2e267e753c725d88a0e9eaa98117127cb21fbf34c0aabb64420bdf)
    Weth[ from ] = [ MevBot, PairB(del) ];	Weth[ to ] = [ PairA(del), MevBot ]
    Edoge[ from ] = [ PairA(del) ];	Edoge[ to ] = [ PairB(del) ]

    经删除后, 剩余: Weth[from] = MevBot; Weth[to] = MevBot . from和to相等, 因此也删除. 这里有2笔swap
    a. eth -> edoge(买入edoge): trader = null.
    b. edoge -> eth(卖出edoge): trader = null.

    如果是 pepe->eth->pepe, 则:
    Pepe[ from ] = MevBot;	Pepe[ to ] = MevBot, 结论是一样的

 5. 对于通缩币, 本质就是一笔正常transfer,他会触发多笔通缩的transfer, 此时需要做的事情便是识别出正常的transfer, 如(https://etherscan.io/tx/0xee9e007a51974b316755a381640d1836466827c8c2f6adfb9222b90a79e32625)
    Reddit[ from ] = [ Operator ]; 	Reddit[ to ] = [ Uni_Pair(del) ]
    Weth[ from ] = [ Pair_A(del) ];	Weth[ to ] = [ Pair_B(del) ]
    Mile[ from ] = [ Pair_B(del), Pair_B(del) ];	Mile[ to ] = [ Addr_A, Addr_B ]

    经提炼后为: Reddit[ from ] = [ Operator ]; 	Mile[ to ] = [ Addr_A, Addr_B ]. 因为from只有一笔, 因此to中只提炼出一笔金额最大的(如果为2笔, 则提炼出2笔)

    假设为多个池子 buy 通缩币 eth(2/3)->uni_deflat; eth(1/3)->sushi_deflat
    Weth[ from ] = [ Operator, Operator ]; Weth[ to ] = [ Uni(del), Sushi(del) ]
    Deflat[ from ] = [ Uni(del), Sushi(del) ];	Deflat[ to ] = [ Receiver, Deflat_A1, Deflat_A2, Receiver, Deflat_A1, Deflat_A2 ]

    经提炼: Weth[ from ] = [ Operator, Operator ]; Deflat[ to ] = [ Receiver, Deflat_A1, Deflat_A2, Receiver, Deflat_A1, Deflat_A2 ]
    tips: 此时由于from和to数组不一样, 确定存在通缩币, 则 累积出最大、第二等依次作为该trader.
    如果他卖或买的池子比例比通缩比例还差距大, 则记录错误(可承受)

    假设为多个池子 sell 通缩币, 则反之 to 数组小, 取to的数组长度作为transfer有效次数

 6. 对于一笔tx多笔买卖的情况, 如用1inch买卖经过多个池子, 如 TokenA->UniEth; Token->SushiEth
    Token[ from ] = [ TFrom, TFrom ];	Token[ to ] = [ Uni(del), Sushi(del) ]
    Weth[ from ] = [ Uni(del), Sushi(del) ];	Weth[ to ] = [ AddressX, AddressX ]

    提炼后为: Token[ from ] = [ TFrom, TFrom ]; Weth[ to ] = [ AddressX, AddressY ]
    这里会出现2笔swap, 均为卖. 根据amount确定 trader(可能不一样)
    token->eth(卖token): trader=token[from] = TFrom

    如果是多笔买, 则提炼后为:
    Weth[ from ] = [ WFrom, WFrom ];	Token[ to ] = [ AdressT, AddressT ]
    同理, 根据金额分别记录 trader
*/
type transferBasic struct {
	Address string
	Amount  float64
}

type transferInfo [2][]transferBasic            // [0][]表示from地址集合, [1][]表示to集合
type tokenTransfer map[string]transferInfo      // key为token, value为transferInfo
type tokensTransferMap map[string]tokenTransfer // 定义数据结构

// basic 判断函数
func contains(slice []transferBasic, item transferBasic) bool {
	for _, s := range slice {
		if s.Address == item.Address {
			return true
		}
	}

	return false
}

// debug
// func printTTM(ttm tokensTransferMap) {
// 	for txhash, tt := range ttm {
// 		fmt.Printf("\n\n\n**TxHash: %v** \n\n", txhash)

// 		for token, ti := range tt {
// 			fmt.Printf("\nToken: %v\n", token)

// 			fmt.Println("Froms: ", ti[0])
// 			fmt.Println("Tos: ", ti[1])
// 		}
// 	}
// }

// 生成 transfer 的 from/to map
func creatTokenTransferMap(transfers []*schema.Transfer, swapContracts map[string]bool) tokensTransferMap {
	ttm := make(tokensTransferMap)

	for _, _transfer := range transfers {
		_, ok := ttm[_transfer.TxHash]
		if !ok { // 先初始化 tokenTransfer
			tt := make(tokenTransfer)
			ttm[_transfer.TxHash] = tt
		}

		fromTo := ttm[_transfer.TxHash][_transfer.Token]

		// 这里要判断是否 地址是否是pair\router或者需要删除的地址, 如果不是, 则append
		_, ok = swapContracts[_transfer.From]
		if !ok {
			fromTo[0] = append(fromTo[0], transferBasic{Address: _transfer.From, Amount: _transfer.Amount}) // 添加 from 地址
		}

		_, ok = swapContracts[_transfer.To]
		if !ok {
			// 对于通缩币来说, 永远是多个to, 但只有一个from, 因此需要将to中的地址根据amount排序...
			fromTo[1] = append(fromTo[1], transferBasic{Address: _transfer.To, Amount: _transfer.Amount}) // 添加 to 地址
		}

		ttm[_transfer.TxHash][_transfer.Token] = fromTo
	}

	// 要继续清理一下数据
	for txhash, tt := range ttm {

		for token, ti := range tt { // ti为 tokenInfo
			// 1. 	如果在单token中, from和to中存在同样地址, 要删除(处理arbi). 因为表示该token最终是没有进出
			// 		当然如果他在一个tx里面有多个重复, 因此需要一个一个的删除
			for i := 0; i < len(ti[0]); {
				if contains(ti[1], ti[0][i]) {
					// 两边同时删除
					for j := 0; j < len(ti[1]); j++ {
						if ti[1][j].Address == ti[0][i].Address {
							ti[1] = append(ti[1][:j], ti[1][j+1:]...)
							break // 只删除一个, 要对等删除
						}
					}

					ti[0] = append(ti[0][:i], ti[0][i+1:]...)
				} else {
					i++
				}
			}

			// 给数组排序, 根据amount大小倒序排序, 方便找出通缩币收款地址
			sort.Slice(ti[0], func(i, j int) bool {
				return ti[0][i].Amount > ti[0][j].Amount
			})

			sort.Slice(ti[1], func(i, j int) bool {
				return ti[1][i].Amount > ti[1][j].Amount
			})

			ttm[txhash][token] = ti
		}

		// 2. 如果经过清洗后在某笔tx中的from和to数组长度不一致, 说明出现了通缩币
		// 暂时反正也只取第一个地址, 因此没必要去删除后边多余地址了

		// var lenFrom, lenTo int
		// for token, ti := range ttm[txhash] { // 不用tt是因为tt可能已经发生了变化
		// 	lenFrom += len(ti[0])
		// 	lenTo += len(ti[1])

		// 	ttm[txhash][token] = ti
		// }

		// 只处理一笔tx中只有一种通缩币的情况, 多个的行为带测试 // todo
		// if lenFrom < lenTo {
		// } else if lenFrom > lenTo {
		// }

	}

	// printTTM(ttm)
	return ttm
}

func UpsertSwapToDB(swaps []*schema.Swap, swapContracts map[string]bool, transfers []*schema.Transfer, mongodb *mongo.Client) {
	handler_Lock.Lock()
	defer handler_Lock.Unlock()

	ttm := creatTokenTransferMap(transfers, swapContracts)

	for _, _swap := range swaps {
		updateTrader(_swap, ttm)
		service_swap.SaveSwapTx(_swap, mongodb)
	}
}

func updateTrader(swap *schema.Swap, ttm tokensTransferMap) {
	if swap.LogNumInHash > config.LogNumTooBigInOneTx {
		swap.Direction = schema.DIRECTION_TOO_COMPLICATED
		return
	}

	if swap.Direction != schema.DIRECTION_BUY_OR_ADD && swap.Direction != schema.DIRECTION_SELL_OR_DECREASE {
		return
	}

	// 过滤出ttm中的from和to地址及token, 原则上只剩余 一个token有 from 地址, 一个token有 to 地址
	// 如果一个tx里面有多笔swap, 则也有可能有多个 from 和 to, 此时根据 swap的买卖方向来决定
	fromTo := ttm[swap.TxHash][swap.MainToken]

	if swap.Direction == schema.DIRECTION_BUY_OR_ADD {
		// 该token是一个买单, 寻找 to 地址
		if len(fromTo[1]) > 0 {
			// 直接取数组第一个, 一笔tx极其复杂的反复买卖同一个token, 还指向不同地址的忽略
			swap.Trader = fromTo[1][0].Address
		}
	} else {
		// sell单, 寻找合法的 from 地址
		if len(fromTo[0]) > 0 {
			// 直接取数组第一个, 一笔tx极其复杂的反复买卖同一个token, 还指向不同地址的忽略
			swap.Trader = fromTo[0][0].Address
		}
	}

	// 判断下如果整个tx中的 from 和 to 数组 长度 均为0, 表示没有进出 token
	// 那修改方向为arbi
	// 如果只是单token那可能长度为0的, 因为可能是 usdt->eth->token
	if swap.Trader == "" {
		var lenFrom, lenTo int
		for _, ti := range ttm[swap.TxHash] {
			lenFrom += len(ti[0])
			lenTo += len(ti[1])
		}

		if lenFrom == 0 && lenTo == 0 {
			swap.Direction = schema.DIRECTION_ARBI
		}
	}

}

/*
@20231122 (未统计合约买卖, 已废弃)
找出真正的受益方或者trader, 默认值为 operator

if buy:

	if Token.transfer.To == operator: (遍历该tx中的transfer记录)
		trader = operator
	else // 如使用1inch挂单, 第三方吃单
		if Token.transfer.To 不为合约的地址:
	trader = To // 第一个就作为To吧, 可能不精确; 另外如果是合约购买的情况，也忽略了

else if sell:

	if Token.transfer.From == operator: (遍历该tx中的transfer记录)
		trader = operator
	else
		if Token.transfer.From 不为合约的地址:
			trader = Token.transfer.From // 如果是操作, 肯定是合约之间流转

如果上述没有找到，报错再分析，同时trader为null，表示作废
*/

// func updateTrader(swap *schema.Swap,) {
// 	if swap.Direction != schema.DIRECTION_BUY_OR_ADD && swap.Direction != schema.DIRECTION_SELL_OR_DECREASE {
// 		return
// 	}

// 	key := fmt.Sprintf("%v_%v", swap.TxHash, swap.MainToken)

// 	transfers, ok := transferMap[key]
// 	if !ok {
// 		return
// 	}

// 	for _, transfer := range transfers {
// 		if swap.Direction == schema.DIRECTION_BUY_OR_ADD {
// 			if transfer.To == swap.Operator {
// 				swap.Trader = swap.Operator
// 				break
// 			}
// 		} else {
// 			if transfer.From == swap.Operator {
// 				swap.Trader = swap.Operator
// 				break
// 			}
// 		}
// 	}

// 	// 如果还没找到, 则遍历该token的transfer地址, 找到金额最大的不为contract的地址
// 	if swap.Trader == "" {
// 		biggestAmount := utils.GetBigIntOrZero(swap.AmountOfMainBig)

// 		// 为了防止通缩币向个人转账导致误判, 这里找出转账金额最大的那个
// 		for _, transfer := range transfers {
// 			address := transfer.To
// 			if swap.Direction == schema.DIRECTION_SELL_OR_DECREASE {
// 				address = transfer.From
// 			}

// 			// 去除一些通用的无效地址
// 			if utils.IsDeadAddress(address) {
// 				continue
// 			}

// 			// 要找到不比交易金额小的 target address
// 			if !chain.IsContract(address) && transfer.AmountBigInt.Cmp(biggestAmount) >= 0 {
// 				swap.Trader = address
// 				biggestAmount = transfer.AmountBigInt // 一直更新until找到最大的
// 			}

// 		}
// 	}

// 	// 如果此时还为空, 则直接认为是trader吧
// 	// 逻辑删除, 因为如果to都没有, 则不是一笔合法买入而可能是一个arbi等, 排除掉影响分析
// 	// if swap.Trader == "" {
// 	// 	swap.Trader = swap.Operator
// 	// }

// }

func UpdateKlineBySwap(swap *schema.Swap, mongodb *mongo.Client) {
	// 数据为最近一周才update kline
	if time.Since(swap.SwapTime).Seconds() < config.SecondsForOneWeek {
		UpdateKlines(swap, mongodb)
	}
}

// 更新 volume 的usd value
// 更新 price 的法币价格
func updateUsdInfo(swap *schema.Swap, mongodb *mongo.Client) {
	// 找到quoteToken, 更新 VolumeInUsd.
	// 如果quoteToken为eth, 则乘以区块中eth价格; 如果为u, 直接加; 其他情况为0
	quoteToken := swap.Token1
	if swap.MainToken == swap.Token1 {
		quoteToken = swap.Token0
	}

	if utils.CheckExistString(quoteToken, config.QuoteUsdCoinList) {
		swap.PriceInUsd = swap.Price
	} else if utils.CheckExistString(quoteToken, config.QuoteEthCoinList) {
		swap.PriceInUsd = swap.Price * swap.CurrentEthPrice
	} else {
		// 从token中取, 还取不到, 那就尴尬一笑
		_token, err := token.GetTokenInfo(swap.MainToken, mongodb)
		if err == nil {
			swap.PriceInUsd = _token.PriceInUsd
		} else {
			// utils.Errorf("[ updateUsdInfo ] Temp error! get price in usd error in swap! token: %v", swap.MainToken)
			swap.PriceInUsd = 0 // 应该报错
		}
	}
	swap.VolumeInUsd = swap.AmountOfMainToken * swap.PriceInUsd
}

func newSwapStruct(block *schema.Block, _log *types.Log, tx *schema.Transaction) *schema.Swap {
	swap := schema.Swap{
		BlockNo:  _log.BlockNumber,
		TxHash:   _log.TxHash.String(),
		Position: _log.TxIndex,

		LogNumInHash: len(tx.Receipt.Logs),

		PairAddr: _log.Address.String(),

		GasPrice: tx.Receipt.EffectiveGasPrice.String(),

		OperatorNonce: tx.OriginTx.Nonce(),

		SwapTime: time.Unix(int64(block.Block.Time()), 0),
	}

	effectiveGasPrice := big.NewInt(int64(tx.Receipt.GasUsed))
	effectiveGasPrice = effectiveGasPrice.Mul(effectiveGasPrice, tx.Receipt.EffectiveGasPrice)
	swap.GasInEth = effectiveGasPrice.String()

	// 解析发送者地址
	sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
	if err != nil {
		utils.Warnf("[ newSwapStruct ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
	} else {
		swap.Operator = sender.String()
	}

	// 获取 token0, token1
	pair, err := pair.GetPairInfoForRead(swap.PairAddr)
	if err == nil {
		swap.Token0 = pair.Token0
		swap.Token1 = pair.Token1
		swap.PairName = pair.PairName
	} else {
		log.Printf("[ newSwapStruct ] wrong pair here. addr: %v, tx: %v\n", swap.PairAddr, swap.TxHash)
		return nil
	}

	swap.LogIndexWithTx = fmt.Sprintf("%s_%d", _log.TxHash.String(), _log.Index)
	swap.CurrentEthPrice = block.EthPrice

	return &swap
}
