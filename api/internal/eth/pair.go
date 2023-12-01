package eth

import (
	"sfilter/api/utils"
	"sfilter/services/pair"

	"github.com/gin-gonic/gin"
)

func GetNewPairs(c *gin.Context) {
	info, err := pair.GetNewPairs(utils.GetMongo())
	if err != nil {
		utils.ResFailure(c, 401, err.Error())
	}

	utils.ResSuccess(c, info)
}
