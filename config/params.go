package config

import (
	"os"
	"sfilter/utils"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	DatabaseName = "deepeye"

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

func init() {

	// 如果本地有env, 则以本地为主
	initEnvConfig()
}

func initEnvConfig() {
	if err := godotenv.Load(".env"); err != nil {
		utils.Warnf("no env file, pass..")
		return
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

}
