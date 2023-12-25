package user

import (
	"context"
	"fmt"
	"net/http"
	lAuth "sfilter/user/auth"
	"sfilter/user/config"
	"sfilter/user/controllers"
	"sfilter/user/models"
	"sfilter/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/middleware"
	"github.com/go-pkgz/auth/token"
	adapter "github.com/gwatts/gin-adapter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 对外的 middleware, 在其他地方需要认证成功的时候使用
var authMiddleware gin.HandlerFunc

func GetUserAuthMiddleware() gin.HandlerFunc {
	return authMiddleware
}

type Server struct {
	DB   *mongo.Client
	Auth *auth.Service
}

func NewServer() *Server {
	clientOptions := options.Client().ApplyURI(config.DbAddress)
	mongodb, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		utils.Fatalf("connect mongo error: ", err)
	}

	_auth := lAuth.NewAuthService(mongodb)
	// 用户名密码注册
	_auth.AddDirectProviderWithUserIDFunc("password", lAuth.PasswordCredCheck(), lAuth.UserIDWithRecordIp())

	return &Server{
		DB:   mongodb,
		Auth: _auth,
	}
}

func AuthAdminMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := token.GetUserInfo(c.Request)
		adminRoleStr := fmt.Sprintf("%d", models.USER_ROLE_LEVEL_ROOT)

		if err == nil && user.Attributes["role"] == adminRoleStr {
			c.Next()
		} else {
			utils.Errorf("[ AuthAdminMiddleWare ] auth failed. err: %v, user role: %v", err, user.Attributes["role"])
			c.JSON(http.StatusUnauthorized, "claim is wrong")
		}

		c.Abort()
	}
}

// 搜了好久资料, 还是这哥们厉害, 很简单的实现了我的想法
// 我只是知道咋回事, 但那个 next 参数想不出来..
func wrapHttpToGinMiddleware(authenticator middleware.Authenticator) gin.HandlerFunc {
	return adapter.Wrap(authenticator.Auth)
}

func Run(r *gin.Engine) {
	server := NewServer()

	// 库内置的routers
	authRoutes, avaRoutes := server.Auth.Handlers()

	authMiddleware = wrapHttpToGinMiddleware(server.Auth.Middleware())
	adminAuthMiddleware := AuthAdminMiddleWare()

	// /auth/login; /auth/logout; /auth/user
	r.Any("/auth/*action", gin.WrapH(authRoutes))  // add auth handlers
	r.Any("/avatar/*action", gin.WrapH(avaRoutes)) // add avatar handler

	// 我们需要的新增 routers, 如注册等
	g := r.Group("/user")
	{
		g.POST("/register", controllers.Register)
		g.GET("/email/code", controllers.SendCode) // 验证码
	}

	gWithAuth := g.Use(authMiddleware)

	{
		// 用户设置
		gWithAuth.GET("/info", controllers.GetUserInfo)
		gWithAuth.POST("/passwd/reset", controllers.ResetPassword)
	}

	{
		// 用户跟踪地址设置
		gWithAuth.GET("/trackaddr", controllers.GetTrackedAddresses) // 获取列表

		gWithAuth.POST("/trackaddr", controllers.CreateTrackedAddress)            // 新增
		gWithAuth.PATCH("/trackaddr/:address", controllers.UpdateTrackedAddress)  // 修改
		gWithAuth.DELETE("/trackaddr/:address", controllers.DeleteTrackedAddress) // 删除
	}

	gWithAdminAuth := g.Group("/admin").Use(adminAuthMiddleware)
	{
		gWithAdminAuth.GET("/users", controllers.AdminGetAllUsers)
		gWithAdminAuth.PATCH("/users", controllers.AdminUpdateRole)
	}
}
