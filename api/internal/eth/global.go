package eth

import (
	"sfilter/api/utils"
	"sfilter/services/global"

	"github.com/gin-gonic/gin"
)

func GetGlobalInfo(c *gin.Context) {
	info, err := global.GetGlobalInfo(utils.GetMongo())
	if err != nil {
		utils.ResFailure(c, 401, err.Error())
		return
	}

	utils.ResSuccess(c, info)
}
