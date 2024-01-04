package config

import (
	"os"
	"sfilter/utils"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	MONGO_ADDR       = "mongodb://127.0.0.1:27017"
	DatabaseName     = "deepeye"
	API_AES_DATA_KEY = "deepeye@leafan16"

	WS_ADDR = "ws://127.0.0.1:8546"

	RetriveOldBlockNum         = BlocksPerDay * 3
	GetPriceIntervalForRetrive = 100 // 每隔多少个区块获取一次eth价格

	SwapSaveTime          = int32(SecondsForOneMonth * 3)
	TransferTableSavetime = int32(SecondsForOneMonth * 3)

	Kline1MinTableSaveTime  = int32(SecondsForOneWeek)
	Kline1HourTableSaveTime = int32(SecondsForOneMonth * 6)

	TokenTableSavetime = int32(SecondsForOneYear * 3)

	ApiListenAddrPort = ":10086"
)

// mongodb 查询时的上限值, 超过则吐10000
var COUNT_UPPER_SIZE = int64(10000)

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

	ws_addr := os.Getenv("WS_ADDR")
	if ws_addr != "" {
		utils.Infof("[ init ] Using node ws addr: %v", ws_addr)
		WS_ADDR = ws_addr
	}

	dbName := os.Getenv("DB_NAME")
	if dbName != "" {
		utils.Infof("[ init ] Using db: %v", dbName)
		DatabaseName = dbName
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
		utils.Infof("[ init ] Using API_AES_DATA_KEY: %v", aesKey)
		API_AES_DATA_KEY = aesKey
	}

}
