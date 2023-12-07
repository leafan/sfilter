package config

import "math/big"

// const CREAT_DEBUG = true
// const DEVELOPMENT = false

const CREAT_DEBUG = false
const DEVELOPMENT = true

const SecondsForOneDay = (60 * 60 * 24)
const SecondsForOneWeek = (SecondsForOneDay * 7)
const SecondsForOneMonth = (SecondsForOneDay * 30)
const SecondsForOneYear = (SecondsForOneMonth * 12)

const BlocksPerDay = 250 * 24
const SleepIntervalforRetrive = 100 // 单位ms, 每隔多久取一次区块

// api configure

const ProxyFromIp = "192.168.2.101"

var (
	DatabaseName = "deepeye"

	RetriveOldBlockNum         = BlocksPerDay * 3
	GetPriceIntervalForRetrive = 10 // 每隔多少个区块获取一次eth价格

	SwapSaveTime          = int32(SecondsForOneMonth * 3)
	TransferTableSavetime = int32(SecondsForOneMonth * 3)

	Kline1MinTableSaveTime  = int32(SecondsForOneWeek)
	Kline1HourTableSaveTime = int32(SecondsForOneMonth * 6)

	TokenTableSavetime = int32(SecondsForOneYear * 3)

	ApiListenAddrPort = ":10086"
)

func init() {
	// 重置 参数..
	if DEVELOPMENT {
		DatabaseName = "test"
		RetriveOldBlockNum = 100
		ApiListenAddrPort = ":50086"
	}

	if CREAT_DEBUG {
		DatabaseName = "creat"

		RetriveOldBlockNum = (BlocksPerDay * 30)
		GetPriceIntervalForRetrive = 100
	}
}

const BlockProceededTableName = "block"
const SwapTableName = "swap"
const PairTableName = "pair"
const TokenTableName = "token"
const TransferTableName = "transfer"
const ConfigTableName = "config"

const LiquidityEventTableName = "liquidity"
const LiquidityEventSaveTime = int32(SecondsForOneYear)

const Kline1MinTableName = "kline1min"
const Kline1HourTableName = "kline1hour"

const GlobalTrendTableName = "trend"
const GlobalTrendTableSaveTime = SecondsForOneWeek

const NeverExpireTime = 0

const MaxConcurrentRoutineNums = 10   // 最大并行的协程数, 避免节点扛不住
const GlobalUpdateIntervalBlocks = 10 // 每隔多少个区块update一次全局24h趋势数据

const INFINITE_CHANGE = 1

const MONGO_LIMIT_UPPER = 50
const MONGO_LIMIT_DOWN = 5
const MONGO_PAGE_UPPER = 1000

// mongodb 查询时的上限值, 超过则吐10000
var COUNT_UPPER_SIZE = int64(10000)

var StartRetriveBlock = 0 // 0表示不设置, 该配置用于调试, 一般不用配置

const MONGO_FIND_TIMEOUT = 2
const MONGO_ADDR = "mongodb://127.0.0.1:27017"

const WS_ADDR = "ws://127.0.0.1:8546"

// 用于获取历史高度上的eth价格.. 如果是回溯的时候，10个区块才调用一次即可(特殊处理)
const INFURA_KEY_ADDR = "https://mainnet.infura.io/v3/06a6594cfd1a404591470c2f81a7ac93"

// 固定的quote币种列表
// wbtc去掉, 因为区块中只有eth价格
var QuoteEthCoinList = []string{
	"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // weth
	// "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", // wbtc

}

var QuoteUsdCoinList = []string{
	"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // usdc
	"0xdAC17F958D2ee523a2206206994597C13D831ec7", // usdt
	"0x6B175474E89094C44Da98b954EedeAC495271d0F", // dai
}

const PriceBaseFactor = 1000000000000000000 // price乘以的基数

var BaseFactor1e18 = big.NewInt(1000000000000000000)

// amount 乘以的基数
var AmountBaseFactor1e36 = BaseFactor1e18.Mul(BaseFactor1e18, BaseFactor1e18)
