package config

const BlocksPerDay = 250 * 24

const SleepIntervalforRetrive = 100 // 单位ms, 每隔多久取一次区块

// const (
// 	// const RetriveOldBlockNum = BlocksPerDay * 3
// 	RetriveOldBlockNum = 1000 // 如果要回溯多一些区块, 修改这个字段

// 	GetPriceIntervalForRetrive = 10 // 每隔多少个区块获取一次eth价格

// 	// swap表保存多久, 单位 seconds; 如果要保存久一些, 修改这里
// 	SwapSaveTime = 60 * 60 * 24 * 7 // 7d

// 	DatabaseName = "deepeye"
// )

// for creat

const (
	RetriveOldBlockNum         = BlocksPerDay * 180
	GetPriceIntervalForRetrive = 100
	SwapSaveTime               = 60 * 60 * 24 * 365
	DatabaseName               = "creat"
)

// for creat end..

const BlockProceededTableName = "block"
const SwapTableName = "swap"
const PairTableName = "pair"
const TokenTableName = "token"
const Kline5MinTableName = "kline5min"

// const SwapSaveTime = 10

const NeverExpireTime = 0

const MaxRoutineNums = 10

const WS_ADDR = "ws://127.0.0.1:8546"
const MONGO_ADDR = "mongodb://127.0.0.1:27017"

// 用于获取历史高度上的eth价格.. 如果是回溯的时候，10个区块才调用一次即可(特殊处理)
const INFURA_KEY_ADDR = "https://mainnet.infura.io/v3/06a6594cfd1a404591470c2f81a7ac93"

// 固定的quote币种列表
var QuoteCoinList = []string{
	"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // weth
	"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", // wbtc

}

var UCoinList = []string{
	"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // usdc
	"0xdAC17F958D2ee523a2206206994597C13D831ec7", // usdt
}

const PriceBaseFactor = 1000000000000000000
