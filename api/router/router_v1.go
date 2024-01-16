package router

import (
	"sfilter/api/internal/eth/admin"
	"sfilter/api/router/chain"
	"sfilter/api/utils"
	"sfilter/user"

	"github.com/gin-gonic/gin"
)

func SetUpV1Router(router *gin.Engine) {
	apiV1Group := router.Group("/v1")

	apiV1Group.Use(utils.AuthChainMiddleWare())
	{
		chain.SetupChainRoutes(apiV1Group)
	}

	adminGroup := router.Group("/admin")
	userAuthMiddleware := user.GetUserAuthMiddleware() // 需要先用user auth取到token..
	adminAuthMiddleware := user.GetAdminAuthMiddleware()

	apiV1WithAdminAuth := adminGroup.Use(userAuthMiddleware).Use(adminAuthMiddleware)
	{
		// get deals
		apiV1WithAdminAuth.GET("/deals", admin.AdminGetAllDeals)

	}
}
