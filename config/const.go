package config

import "math/big"

const DEVELOPMENT = true

const SecondsForOneDay = (60 * 60 * 24)
const SecondsForOneWeek = (SecondsForOneDay * 7)
const SecondsForOneMonth = (SecondsForOneDay * 30)
const SecondsForOneYear = (SecondsForOneMonth * 12)

const BlocksPerDay = 250 * 24
const SleepIntervalforRetrive = 100 // 单位ms, 每隔多久取一次区块

// api configure

const ProxyFromIp = "192.168.2.101"

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

const MONGO_FIND_TIMEOUT = 2
const MONGO_ADDR = "mongodb://127.0.0.1:27017"

const WS_ADDR = "ws://127.0.0.1:8546"

const USER_DB_FILE = "/data/sqlite/user.db"

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

var BaseFactor1e18 = big.NewInt(1000000000000000000)
