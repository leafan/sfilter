package main

import (
	"context"
	"net/http"
	"sfilter/api"
	"sfilter/config"
	"sfilter/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-gomail/gomail"
	"github.com/jinzhu/gorm"
	"github.com/qor/auth"
	"github.com/qor/auth/auth_identity"
	"github.com/qor/auth/providers/password"
	"github.com/qor/mailer"
	"github.com/qor/mailer/gomailer"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
	mongodb, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		utils.Fatalf("connect mongo error: ", err)
	}

	r := gin.New()

	// initUserServer(r)

	server := api.NewServer(r, mongodb)
	server.Route()

	utils.Debugf("run server now...")
	server.Run(config.ApiListenAddrPort)
}

func initUserServer(r *gin.Engine) {
	userdb, err := gorm.Open("sqlite3", config.USER_DB_FILE)
	if err != nil {
		utils.Fatalf("connect userdb error: ", err)
	}

	userHandler := getUserServer(userdb)
	r.Any("/auth/*action", gin.WrapH(userHandler))
}

func getUserServer(userdb *gorm.DB) *http.ServeMux {
	auth := initAuth(userdb)

	mux := http.NewServeMux()
	mux.Handle("/auth/", auth.NewServeMux())

	return mux
}

func initAuth(userdb *gorm.DB) *auth.Auth {
	// Migrate AuthIdentity model, AuthIdentity will be used to save auth info, like username/password, oauth token, you could change that.
	userdb.AutoMigrate(&auth_identity.AuthIdentity{})

	Auth := auth.New(&auth.Config{
		DB: userdb,
	})

	// Register Auth providers
	// Allow use username/password
	Auth.RegisterProvider(password.New(&password.Config{}))

	// 注册邮件服务
	dialer := gomail.NewDialer(config.SMTP_HOST, config.SMTP_PORT, config.SMTP_USER, config.SMTP_PASSWORD)
	sender, err := dialer.Dial()

	if err == nil {
		Mailer := mailer.New(&mailer.Config{
			Sender: gomailer.New(&gomailer.Config{Sender: sender}),
		})
		Auth.Mailer = Mailer
	} else {
		utils.Errorf("new mail sender err: ", err)
	}

	return Auth
}
