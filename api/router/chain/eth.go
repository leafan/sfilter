package chain

import (
	"log"
	"sfilter/api/internal/chain"
	"sfilter/api/utils"

	"github.com/gin-gonic/gin"
)

func setupEthRoutes(parentGroup *gin.RouterGroup) {
	ethGroup := parentGroup.Group("/eth")

	getMiddleware := utils.AuthNothingMiddleWare()
	ethGroup.Use(getMiddleware)
	{
		utils.EmptyMigrate(ethGroup)

		ethGroup.GET("/global", chain.GetGlobalInfo)
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
