package chain

import (
	"sfilter/api/internal/eth"
	"sfilter/api/internal/eth/admin"
	"sfilter/api/internal/eth/encrypt"
	"sfilter/api/internal/eth/user"
	"sfilter/api/utils"
	guser "sfilter/user"

	"github.com/gin-gonic/gin"
)

func setupEthRoutes(parentGroup *gin.RouterGroup) {
	ethGroup := parentGroup.Group("/:chain")

	getMiddleware := utils.AuthNothingMiddleWare()
	noAuthGroup := ethGroup.Use(getMiddleware)

	{
		utils.EmptyMigrate(ethGroup)

		// overview
		noAuthGroup.GET("/global", eth.GetGlobalInfo)

		// pair
		{
			noAuthGroup.GET("/pair", eth.GetPair)

			// 优秀的pair列表. 默认筛选最近30天add的池子(是个选项); 按24h成交数排序
			// 支持其他方式排序;
			noAuthGroup.GET("/hotpair", eth.GetHotPairs)

			// 按24h交易数排序, 且最近7天内新添加池子的pair
			noAuthGroup.GET("/hotnewpair", eth.GetHotPairs)
		}

		// liquidity
		{
			// 所有的最新 add/remove liquidity 集合, 支持 token, pair, operator 等条件查询
			noAuthGroup.GET("/liquidity", eth.GetLiquidityEvent)
		}

		// trend
		{
			noAuthGroup.GET("/pairtrend", eth.GetPriceAndTxTrends)
		}

		// 简单加密数据
		{
			// 所有的最新swaps集合, 支持 token, operator, trader 等条件查询
			// 只支持查询最近1个月的数据
			noAuthGroup.GET("/swaps", encrypt.GetSwapEvents)

			// 所有的最新transfer集合, 支持 token, operator 等条件查询
			noAuthGroup.GET("/transfers", encrypt.GetTransferEvents)

			// 所有的铭文查询
			noAuthGroup.GET("/facets", encrypt.GetFacets)

			// 所有的facet查询
			noAuthGroup.GET("/inscriptions", encrypt.GetInscriptions)

		}

	}

	// apiKey
	apiKeyMiddleware := guser.GetApiKeyAuthMiddleware()
	apiKeyGroup := ethGroup.Group("/apikey").Use(apiKeyMiddleware)
	{
		apiKeyGroup.GET("/trackswaps", user.GetTrackSwaps)
		apiKeyGroup.GET("/pairs", eth.GetHotPairs)

		// for partners
		apiKeyGroup.GET("/hbpair", admin.GetHotBigPairs)
	}

	// auth user etc..
	authMiddleware := guser.GetUserAuthMiddleware()
	authGroup := ethGroup.Use(authMiddleware)
	{
		authGroup.GET("/trackswaps", user.GetTrackSwaps)
	}

}
