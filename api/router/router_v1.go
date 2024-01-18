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

	userAuthMiddleware := user.GetUserAuthMiddleware() // 需要先用user auth取到token..

	partnerGroup := router.Group("/partner")
	partnerAuthMiddleware := user.GetPartnerAuthMiddleware()

	partnerGroupWithAuth := partnerGroup.Use(userAuthMiddleware).Use(partnerAuthMiddleware)
	{
		// get deals
		partnerGroupWithAuth.GET("/deals", admin.AdminGetDeals)
		partnerGroupWithAuth.GET("/wisers", admin.AdminGetWisers)
	}

	adminGroup := router.Group("/admin")
	adminAuthMiddleware := user.GetAdminAuthMiddleware()

	adminGroupWithAuth := adminGroup.Use(userAuthMiddleware).Use(adminAuthMiddleware)
	{
		// get deals
		adminGroupWithAuth.GET("/deals", admin.AdminGetDeals)

	}

}
