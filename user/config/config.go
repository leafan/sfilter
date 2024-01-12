package config

import (
	"os"
	"time"

	globalConfig "sfilter/config"
	"sfilter/utils"

	"github.com/joho/godotenv"
)

var (
	Issuer       = "deepeye.cc"
	JWTHeaderKey = "Authorization"
	APIURL       = "https://api.deepeye.cc"

	TokenDuration  = 24 * 7 * time.Hour // jwt有效期
	CookieDuration = 24 * 7 * time.Hour // cookie有效期

	DbAddress    = globalConfig.MONGO_ADDR
	DatabaseName = "user"

	UserTableName = "users"
	UserWatchList = "watchlist"

	VerifyCodeTableName = "vcode"
	VerifyCodeMaxNum    = 3 // 某一段时间之内某ip允许的最大请求验证码数

	LoginHistoryTableName = "logins"
	TrackAddressTableName = "trackaddr"
)

var (
	JWTSecret     = ""
	IPINFO_APIKEY = "4af0cf25375a5d" // leafan.chan google账号注册
)

var (
	// aws  key
	AWS_KEY_ID     = ""
	AWS_SECRET_KEY = ""
	AWS_REGION     = "ap-southeast-1"
)

func init() {
	// 如果本地有env, 则以本地为主
	initEnvConfig()
}

func initEnvConfig() {
	if err := godotenv.Load(".env"); err != nil {
		utils.Warnf("no env file, pass..")
		return
	}

	if os.Getenv("AWS_KEY_ID") != "" {
		AWS_KEY_ID = os.Getenv("AWS_KEY_ID")
	}
	if os.Getenv("AWS_SECRET_KEY") != "" {
		AWS_SECRET_KEY = os.Getenv("AWS_SECRET_KEY")
	}
	if os.Getenv("AWS_REGION") != "" {
		AWS_REGION = os.Getenv("AWS_REGION")
	}
	if os.Getenv("JWT_SECRET") != "" {
		JWTSecret = os.Getenv("JWT_SECRET")
	}

	if os.Getenv("DATABASE_USER") != "" {
		DatabaseName = os.Getenv("DATABASE_USER")
	}

}
