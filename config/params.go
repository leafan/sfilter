package config

import (
	"os"
	"sfilter/utils"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	DevelopmentMode = false

	MONGO_ADDR   = "mongodb://127.0.0.1:27017"
	DatabaseName = "deepeye"

	GlobalDatabaseName = "" //  全局配置信息获取表, 如获取 pair 和 token等
	API_AES_DATA_KEY   = "deepeye@leafan16"

	// 用于获取历史高度上的eth价格.. 如果是回溯的时候，10个区块才调用一次即可(特殊处理)
	INFURA_API_KEY = "https://mainnet.infura.io/v3/06a6594cfd1a404591470c2f81a7ac93"

	WS_ADDR = "ws://127.0.0.1:8546"

	RetriveOldBlockNum         = BlocksPerDay * 3
	GetPriceIntervalForRetrive = 20 // 每隔多少个区块获取一次eth价格

	SwapSaveTime          = int32(SecondsForOneMonth * 3)
	TransferTableSavetime = int32(SecondsForOneMonth * 3)

	Kline1MinTableSaveTime  = int32(SecondsForOneWeek * 3)
	Kline1HourTableSaveTime = int32(SecondsForOneMonth * 6)

	TokenTableSavetime = int32(SecondsForOneYear * 3)

	ApiListenAddrPort = ":10086"
)

// mongodb 查询时的上限值
var COUNT_UPPER_SIZE = int64(10000)
var SELECT_UPPER_SIZE = int64(10000)

// 定义facet项目名称, 目前只知道 facetswap..
func GetFacetProjectName(addr string) string {
	if addr == "0x00000000000000000000000000000000000FacE7" {
		return "Facet"
	}

	return "Unknown"
}

func init() {

	// 如果本地有env, 则以本地为主
	initEnvConfig()
}

func initEnvConfig() {
	if err := godotenv.Load(".env"); err != nil {
		utils.Warnf("no env file, pass..")
		return
	}

	mongoAddr := os.Getenv("MONGO_ADDR")
	if mongoAddr != "" {
		utils.Infof("[ init ] Using mongo addr: %v", mongoAddr)
		MONGO_ADDR = mongoAddr
	}

	infura_apikey := os.Getenv("INFURA_API_KEY")
	if infura_apikey != "" {
		utils.Infof("[ init ] Using infura_apikey: %v", infura_apikey)
		INFURA_API_KEY = infura_apikey
	}

	ws_addr := os.Getenv("WS_ADDR")
	if ws_addr != "" {
		utils.Infof("[ init ] Using ws addr: %v", ws_addr)
		WS_ADDR = ws_addr
	}

	dbName := os.Getenv("DB_NAME")
	if dbName != "" {
		utils.Infof("[ init ] Using db: %v", dbName)
		DatabaseName = dbName
	}
	globalDbName := os.Getenv("GLOBAL_DB_NAME")
	if globalDbName != "" {
		utils.Infof("[ init ] Using global db: %v", globalDbName)
		GlobalDatabaseName = globalDbName
	}

	retriveNum := os.Getenv("RETRIVE_OLD_BLOCK_NUM")
	if retriveNum != "" {
		tmpNum, err := strconv.Atoi(retriveNum)
		if err != nil {
			utils.Fatalf("Wrong param of RETRIVE_OLD_BLOCK_NUM: %v", retriveNum)
		}

		utils.Infof("[ init ] Using RetriveOldBlockNum: %v", tmpNum)
		RetriveOldBlockNum = tmpNum
	}

	listenAddr := os.Getenv("API_LISTEN_PORT")
	if listenAddr != "" {
		utils.Infof("[ init ] Using ApiListenAddrPort: %v", listenAddr)
		ApiListenAddrPort = listenAddr
	}

	aesKey := os.Getenv("API_AES_DATA_KEY")
	if aesKey != "" {
		// utils.Infof("[ init ] Using API_AES_DATA_KEY: %v", aesKey)
		API_AES_DATA_KEY = aesKey
	}

}
