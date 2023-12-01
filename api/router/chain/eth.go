package chain

import (
	"log"
	"sfilter/api/internal/eth"
	"sfilter/api/utils"

	"github.com/gin-gonic/gin"
)

func setupEthRoutes(parentGroup *gin.RouterGroup) {
	ethGroup := parentGroup.Group("/eth")

	getMiddleware := utils.AuthNothingMiddleWare()
	ethGroup.Use(getMiddleware)
	{
		utils.EmptyMigrate(ethGroup)

		// overview
		ethGroup.GET("/global", eth.GetGlobalInfo)

		// pair
		{
			// 最新加流动性的pair列表，不支持其他方式排序, 单纯的展示
			// 支持按 token、pair 地址查询
			ethGroup.GET("/newpair", eth.GetNewPairs)

			// 优秀的pair列表. 默认筛选最近30天add的池子(是个选项); 按24h成交数排序
			// 支持其他方式排序; 同时还有一个24h交易额最少大于10000u的选项
			ethGroup.GET("/goodpair", eth.GetNewPairs)
		}

		// swap
		{
			// 所有的最新swaps集合, 支持 token, operator, trader 等条件查询
			// 只支持查询最近1个月的数据
			ethGroup.GET("/swaps", eth.GetNewPairs)
		}

		// transfer
		{
			// 所有的最新transfer集合, 支持 token, operator 等条件查询
			ethGroup.GET("/transfers", eth.GetNewPairs)
		}

		// liquidity
		{
			// 所有的最新 add/remove liquidity 集合, 支持 token, pair, operator 等条件查询
			ethGroup.GET("/liquidity", eth.GetLiquidityEvent)
		}

		// block
		{
			// 所有的最新 block 集合, 支持 token, pair, operator 等条件查询
			ethGroup.GET("/blocks", eth.GetNewPairs)

			// 某区块下的所有信息, 含swaps集合
			ethGroup.GET("/block/:no", eth.GetNewPairs)
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
