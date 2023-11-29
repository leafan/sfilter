package chain

import "github.com/gin-gonic/gin"

func SetupChainRoutes(parentGroup *gin.RouterGroup) {
	setupEthRoutes(parentGroup)

	// etc..
}
