package chain

import (
	"log"
	"sfilter/api/internal/eth"
	"sfilter/api/utils"

	"github.com/gin-gonic/gin"
)

func setupEthRoutes(parentGroup *gin.RouterGroup) {
	ethGroup := parentGroup.Group("/:chain")

	getMiddleware := utils.AuthNothingMiddleWare()
	ethGroup.Use(getMiddleware)
	{
		utils.EmptyMigrate(ethGroup)

		// overview
		ethGroup.GET("/global", eth.GetGlobalInfo)

		// pair
		{
			ethGroup.GET("/pair", eth.GetPair)

			// 优秀的pair列表. 默认筛选最近30天add的池子(是个选项); 按24h成交数排序
			// 支持其他方式排序;
			ethGroup.GET("/hotpair", eth.GetHotPairs)

			// 按24h交易数排序, 且最近7天内新添加池子的pair
			ethGroup.GET("/hotnewpair", eth.GetHotPairs)
		}

		// swap
		{
			// 所有的最新swaps集合, 支持 token, operator, trader 等条件查询
			// 只支持查询最近1个月的数据
			ethGroup.GET("/swaps", eth.GetSwapEvents)
		}

		// transfer
		{
			// 所有的最新transfer集合, 支持 token, operator 等条件查询
			ethGroup.GET("/transfers", eth.GetTransferEvents)
		}

		// liquidity
		{
			// 所有的最新 add/remove liquidity 集合, 支持 token, pair, operator 等条件查询
			ethGroup.GET("/liquidity", eth.GetLiquidityEvent)
		}

		// trend
		{
			ethGroup.GET("/pairtrend", eth.GetPriceAndTxTrends)
		}
	}

	// post etc..
	postMiddleware := utils.AuthNothingMiddleWare()
	ethGroup.Use(postMiddleware)
	{
		ethGroup.POST("/test", func(context *gin.Context) {
			go func() {
				log.Printf("[ ethGroup ] post test..\n")
			}()
			utils.ResSuccess(context, gin.H{"message": "post test successful"})
		})
	}
}
