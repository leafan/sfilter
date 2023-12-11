package config

import (
	"log"
	"os"
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

var StartRetriveBlock = 0 // 0表示不设置, 该配置用于调试, 一般不用配置

func init() {

	// 如果本地有env, 则以本地为主
	initEnvConfig()
}

func initEnvConfig() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no env file, pass..")
		return
	}

	dbName := os.Getenv("DB_NAME")
	if dbName != "" {
		log.Println("[ init ] Using db: ", dbName)
		DatabaseName = dbName
	}

	retriveNum := os.Getenv("RETRIVE_OLD_BLOCK_NUM")
	if retriveNum != "" {
		tmpNum, err := strconv.Atoi(retriveNum)
		if err != nil {
			log.Fatal("Wrong param of RETRIVE_OLD_BLOCK_NUM: ", retriveNum)
		}

		log.Println("[ init ] Using RetriveOldBlockNum: ", tmpNum)
		RetriveOldBlockNum = tmpNum
	}

	listenAddr := os.Getenv("API_LISTEN_PORT")
	if listenAddr != "" {
		log.Println("[ init ] Using ApiListenAddrPort: ", listenAddr)
		ApiListenAddrPort = listenAddr
	}
}
