package eth

import (
	"sfilter/api/utils"
	"sfilter/services/liquidity"

	"github.com/gin-gonic/gin"
)

func GetLiquidityEvent(c *gin.Context) {
	info, err := liquidity.GetLiquidityEvents(1, 10, utils.GetMongo())
	if err != nil {
		utils.ResFailure(c, 401, err.Error())
	}

	utils.ResSuccess(c, info)
}
