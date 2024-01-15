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

	adminAuthMiddleware := user.GetAdminAuthMiddleware()
	apiV1WithAdminAuth := apiV1Group.Group("/admin").Use(adminAuthMiddleware)
	{
		// get deals
		apiV1WithAdminAuth.GET("/deals", admin.AdminGetAllDeals)

	}
}
