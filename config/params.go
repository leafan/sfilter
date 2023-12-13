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

	// user configure
	USER_DB_FILE = "/data/sqlite/user.db"

	// mail
	SMTP_HOST     = "127.0.0.1"
	SMTP_PORT     = 10079
	SMTP_USER     = "abc"
	SMTP_PASSWORD = "ddss"
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

	// user config
	smtp_host := os.Getenv("SMTP_HOST")
	if listenAddr != "" {
		utils.Infof("[ init ] Using SMTP_HOST: %v", smtp_host)

		SMTP_HOST = smtp_host
	}

	user_dbfile := os.Getenv("USER_DB_FILE")
	if listenAddr != "" {
		utils.Infof("[ init ] Using USER_DB_FILE: %v", user_dbfile)

		USER_DB_FILE = user_dbfile
	}

	smtp_port := os.Getenv("SMTP_PORT")
	if smtp_port != "" {
		tmpNum, err := strconv.Atoi(smtp_port)
		if err != nil {
			utils.Fatalf("Wrong param of SMTP_PORT: %v", smtp_port)
		}

		utils.Infof("[ init ] Using SMTP_PORT: %v", tmpNum)
		SMTP_PORT = tmpNum
	}

	smtp_user := os.Getenv("SMTP_USER")
	if listenAddr != "" {
		utils.Infof("[ init ] Using SMTP_USER: %v", smtp_user)

		SMTP_USER = smtp_user
	}

	smtp_pass := os.Getenv("SMTP_PASSWORD")
	if listenAddr != "" {
		utils.Infof("[ init ] Using SMTP_PASSWORD: %v", smtp_pass)

		SMTP_PASSWORD = smtp_pass
	}

}
