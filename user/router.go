package user

import (
	"context"
	"fmt"
	"net/http"
	lAuth "sfilter/user/auth"
	"sfilter/user/config"
	"sfilter/user/controllers"
	"sfilter/user/controllers/admin"
	"sfilter/user/models"
	"sfilter/utils"
	"strconv"

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
var apiKeyAuthMiddleware gin.HandlerFunc
var adminAuthMiddleware gin.HandlerFunc
var partnerAuthMiddleware gin.HandlerFunc

func GetUserAuthMiddleware() gin.HandlerFunc {
	return authMiddleware
}

func GetApiKeyAuthMiddleware() gin.HandlerFunc {
	return apiKeyAuthMiddleware
}

func GetAdminAuthMiddleware() gin.HandlerFunc {
	return adminAuthMiddleware
}

func GetPartnerAuthMiddleware() gin.HandlerFunc {
	return partnerAuthMiddleware
}

type Server struct {
	DB   *mongo.Client
	Auth *auth.Service
}

func NewServer() *Server {
	clientOptions := options.Client().ApplyURI(config.DbAddress)
	mongodb, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		utils.Fatalf("connect mongo error: %v", err)
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
			utils.Errorf("[ AuthAdminMiddleWare ] someone visits admin page but failed! err: %v, user role: %v", err, user.Attributes["role"])
			c.JSON(http.StatusUnauthorized, "claim is wrong")

			c.Abort()
		}
	}
}

func AuthPartnerMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := token.GetUserInfo(c.Request)
		if err == nil {
			userRoleInt, err := strconv.Atoi(user.Attributes["role"].(string))

			if err == nil && userRoleInt >= models.USER_ROLE_LEVEL_PARTNER {
				c.Next()
				return
			}
		}

		utils.Errorf("[ AuthPartnerMiddleWare ] someone visits partner page but failed! err: %v, user role: %v", err, user.Attributes["role"])
		c.JSON(http.StatusUnauthorized, "claim is wrong")

		c.Abort()
	}
}

// AuthAPIKeyMiddleware 是一个用于 API Key 认证的中间件
func AuthAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取传入的 API 密钥
		clientAPIKey := c.Query("apikey")
		user, err := models.GetUserInfoByAPIKey(clientAPIKey)

		// 检查传入的 API 密钥是否与预期的 API 密钥匹配
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})

			c.Abort()
			return
		}

		// 将用户信息存储在请求上下文中
		c.Set("user", user.Username)

		// 如果 API 密钥有效，允许请求继续处理
		c.Next()
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

	apiKeyAuthMiddleware = AuthAPIKeyMiddleware()
	adminAuthMiddleware = AuthAdminMiddleWare()
	partnerAuthMiddleware = AuthPartnerMiddleWare()

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

		// 用户跟踪地址设置
		gWithAuth.GET("/trackaddr", controllers.GetTrackedAddresses) // 获取列表

		gWithAuth.POST("/trackaddr", controllers.CreateTrackedAddress)            // 新增
		gWithAuth.PATCH("/trackaddr/:address", controllers.UpdateTrackedAddress)  // 修改
		gWithAuth.DELETE("/trackaddr/:address", controllers.DeleteTrackedAddress) // 删除
	}

	gWithAdminAuth := g.Group("/admin").Use(adminAuthMiddleware)
	{
		// users
		gWithAdminAuth.GET("/users", admin.AdminGetAllUsers)
		gWithAdminAuth.PATCH("/users", admin.AdminUpdateRole)

		// track address
		gWithAdminAuth.GET("/trackaddrs", admin.AdminGetAllTrackedAddress)

	}

}
