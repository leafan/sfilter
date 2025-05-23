package config

import (
	"math/big"
	"time"
)

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
const RouterTableName = "router"
const SpecialAddressTableName = "saddress"
const ConfigTableName = "config"

const LiquidityEventTableName = "liquidity"
const LiquidityEventSaveTime = int32(SecondsForOneYear)

const Kline1MinTableName = "kline1min"
const Kline1HourTableName = "kline1hour"

const GlobalTrendTableName = "trend"
const GlobalTrendTableSaveTime = SecondsForOneWeek

const TrackSwapTableName = "trackswap"
const TrackSwapTableSaveTime = SecondsForOneMonth

const UpdateTrackAddressInterval = 1 * time.Minute // 定时更新用户跟踪地址列表
const CheckTrackAddressInterval = 60 * time.Minute // 定时检查用户记录是否到达上限

const MaxConcurrentRoutineNums = 10   // 最大并行的协程数, 避免节点扛不住
const GlobalUpdateIntervalBlocks = 10 // 每隔多少个区块update一次全局24h趋势数据

const MONGO_LIMIT_UPPER = 50          // 普通用户一页的limit大小
const MONGO_APIKEY_LIMIT_UPPER = 1000 // apikey 用户一页的limit大小
const MONGO_LIMIT_DOWN = 5
const MONGO_PAGE_UPPER = 1000

const MONGO_FIND_TIMEOUT = 5

var BaseFactor1e18 = big.NewInt(1000000000000000000)

// 合法的chain, 只支持这些链
var ValidChains = []string{
	"eth",
	"avax",
	// "optimism",
	// "arbi",
	// "bsc",
}

// facet 定义
const FacetTableName = "facet"
const FacetSaveTime = int32(SecondsForOneMonth)

const InscriptionTableName = "inscription"
const InscriptionSaveTime = int32(SecondsForOneMonth)

// wiser 定义
const WiserTableName = "wiser"

const BiDealTableName = "deals"
const BiDealSaveTime = int32(SecondsForOneYear)

const BiTradeTableName = "trades"

const HotPairRankTableName = "hrank"

// chain 定义
const Quoter_Contract_Address = "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"

const Hacker_Check_Contract_Address = "0x595cD1294e0bdd7897Ac2ABcA6d64A676c3855E5"

const LogNumTooBigInOneTx = 100
