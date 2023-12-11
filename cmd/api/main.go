package main

import (
	"context"
	"log"
	"net/http"
	"sfilter/api"
	"sfilter/config"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/qor/auth"
	"github.com/qor/auth/auth_identity"
	"github.com/qor/auth/providers/github"
	"github.com/qor/auth/providers/password"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
	mongodb, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("connect mongo error: ", err)
	}

	r := gin.New()

	// initUserServer(r)

	server := api.NewServer(r, mongodb)
	server.Route()

	log.Println("run server now...")
	server.Run(config.ApiListenAddrPort)
}

func initUserServer(r *gin.Engine) {
	userdb, err := gorm.Open("sqlite3", config.USER_DB_FILE)
	if err != nil {
		log.Fatal("connect userdb error: ", err)
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

	// Allow use Github
	Auth.RegisterProvider(github.New(&github.Config{
		ClientID:     "github client id",
		ClientSecret: "github client secret",
	}))

	// dialer := gomail.NewDialer(Config.SMTP.Host, Config.SMTP.Port, Config.SMTP.User, Config.SMTP.Password)
	// sender, err := dialer.Dial()

	// Mailer = mailer.New(&mailer.Config{
	// 	Sender: gomailer.New(&gomailer.Config{Sender: sender}),
	// })
	// Auth.Mailer = Mailer

	return Auth
}
