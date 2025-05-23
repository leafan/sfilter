package encrypt

import (
	"sfilter/api/utils"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/swap"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetSwapEvents(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	options, err := parseSwapOptions(c)
	if err != nil {
		return
	}

	filter := parseSwapFilterOptions(c)

	info, count, err := swap.GetSwapEvents(options, filter, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	data := struct {
		List  []schema.Swap `json:"list"`
		Count int64         `json:"count"`
	}{
		List:  info,
		Count: count,
	}

	// 加密
	enc, err := utils.AesEncrypt(data, config.API_AES_DATA_KEY)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}
	utils.ResSuccess(c, enc)
}

func parseSwapOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "swapTime", Value: -1}})

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")
	if sortBy != "" && (sortBy == "amountOfMainToken" || sortBy == "volumeInUsd") {
		key = sortBy

		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}
		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseSwapFilterOptions(c *gin.Context) *primitive.M {
	filter := bson.M{}

	blockNo := c.DefaultQuery("blockNo", "")
	blockNoInt, err := strconv.Atoi(blockNo)
	if err == nil {
		filter["blockNo"] = blockNoInt
	}

	direction := c.DefaultQuery("direction", "0")
	directionInt, err := strconv.Atoi(direction)
	if err == nil && (directionInt == 1 || directionInt == 2) {
		filter["direction"] = directionInt
	}

	trader := c.DefaultQuery("trader", "")
	if trader != "" && utils.IsValidEthereumAddress(trader) {
		filter["trader"] = trader
	}

	pairAddr := c.DefaultQuery("pairAddr", "")
	if pairAddr != "" && utils.IsValidEthereumAddress(pairAddr) {
		filter["pairAddr"] = pairAddr
	}

	// 根据token地址查询, 可能为token0或token1
	token := c.DefaultQuery("_token", "")
	if token != "" && utils.IsValidEthereumAddress(token) {
		filter["$or"] = []bson.M{
			{"token0": token},
			{"token1": token},
			{"pairAddr": token},
		}
	}

	recentdays := c.DefaultQuery("recentdays", "1")
	recentdaysInt, err := strconv.Atoi(recentdays)
	if err == nil {
		if recentdaysInt < 1 {
			recentdaysInt = 1
		} else if recentdaysInt > 30 {
			recentdaysInt = 30
		}

		date := time.Now().Add(-time.Duration(recentdaysInt) * time.Hour * 24)
		filter["swapTime"] = bson.M{
			"$gte": date,
		}
	}

	return &filter
}
