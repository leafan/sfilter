package chain

import (
	"sfilter/api/utils"
	"sfilter/services/global"

	"github.com/gin-gonic/gin"
)

func GetGlobalInfo(c *gin.Context) {
	info, err := global.GetGlobalInfo(utils.GetMongo())
	if err != nil {
		utils.ResFailure(c, 401, err.Error())
	}

	utils.ResSuccess(c, info)
}
