package router

import (
	"sfilter/api/router/chain"
	"sfilter/api/utils"

	"github.com/gin-gonic/gin"
)

func SetUpV1Router(router *gin.Engine) {
	apiV1Group := router.Group("/v1")

	apiV1Group.Use(utils.AuthNothingMiddleWare())
	{
		chain.SetupChainRoutes(apiV1Group)
	}
}
