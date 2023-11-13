package config

const BlocksPerDay = 288 * 24 // 一天预计的blocks数, 10-15s一个区块
// const RetriveOldBlockNum = BlocksPerDay * 3

const RetriveOldBlockNum = 100

const DatabaseName = "deepeye"

const BlockProceededTableName = "block"
const SwapTableName = "swap"
const PairTableName = "pair"
const TokenTableName = "token"

const SwapSaveTime = 60 * 60 * 24 * 7 // 7d
// const SwapSaveTime = 10

// const BlockProceededSaveTime = 60 * 60 * 24 * 30 // 30d
const BlockProceededSaveTime = 0

const NeverExpireTime = 0

const WS_ADDR = "ws://127.0.0.1:8546"
const MONGO_ADDR = "mongodb://127.0.0.1:27017"

// 用于获取历史高度上的eth价格.. 如果是回溯的时候，10个区块才调用一次即可(特殊处理)
const INFURA_KEY_ADDR = "https://mainnet.infura.io/v3/5d6ed0aa25824d2096f7a2827e36105c"

// 固定的quote币种列表
var QuoteCoinList = []string{
	"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // weth
	"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // usdc
}
