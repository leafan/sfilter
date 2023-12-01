package api

import (
	"log"
	"sfilter/api/router"
	"sfilter/api/utils"
    "sfilter/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	Engine *gin.Engine
	Mongo  *mongo.Client
}

func NewServer(engine *gin.Engine, db *mongo.Client) *Server {
	utils.InitMongo(db)

	return &Server{
		Engine: engine,
	}
}

func (server *Server) Run(port string) {
	err := server.Engine.Run(port)
	if err != nil {
		log.Fatal("engine run failed. err: ", err)
		return
	}
}

func (server *Server) Route() {
	// 配置整体的 middleware
	server.configureMiddleware()

	server.Engine.GET("/ping", ping)

	// setup router
	router.SetUpV1Router(server.Engine)
}

func (server *Server) configureMiddleware() {
	server.Engine.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/ping", "/metrics"),
		gin.Recovery(),
	)

	server.Engine.Use(utils.Cors())

    whitelist := make(map[string]bool)
    whitelist[config.ProxyFromIp] = true

    // 设置白名单访问
    server.Engine.Use(utils.IPWhiteList(whitelist))
}

func ping(c *gin.Context) {
	utils.ResSuccess(c, gin.H{"message": "pong"})
}
