package config

import (
	"os"
	"time"

	globalConfig "sfilter/config"
	"sfilter/utils"

	"github.com/joho/godotenv"
)

var (
	MailerApiKey = "mlsn.ed06601d2db7fa3c62e77fc72e69d72a33a7a59e0238f2240d8d96fbf0d38a3a"

	Issuer    = "deepeye.cc"
	JWTSecret = "jwt.I*d2&s@deepeye.cc"
	APIURL    = "https://api.deepeye.cc"

	TokenDuration  = 600000 * time.Minute // jwt有效期
	CookieDuration = 24 * time.Hour       // cookie有效期

	DbAddress    = globalConfig.MONGO_ADDR
	DatabaseName = globalConfig.DatabaseName

	UserDbName = "user"
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

	if os.Getenv("MAILERSEND_API_KEY") != "" {
		MailerApiKey = os.Getenv("MAILERSEND_API_KEY")
	}
}
