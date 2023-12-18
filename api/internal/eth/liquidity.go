package eth

import (
	"sfilter/api/utils"
	"sfilter/schema"
	"sfilter/services/liquidity"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetLiquidityEvent(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	options, err := parseLiquidityOptions(c)
	if err != nil {
		return
	}

	filter := parseLiquidityFilter(c)

	info, count, err := liquidity.GetLiquidityEvents(options, filter, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	data := struct {
		List  []schema.LiquidityEvent `json:"list"`
		Count int64                   `json:"count"`
	}{
		List:  info,
		Count: count,
	}

	utils.ResSuccess(c, data)
}

func parseLiquidityOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")
	if sortBy == "amountInUsd" {
		key = sortBy
		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}

		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseLiquidityFilter(c *gin.Context) *primitive.M {
	filter := bson.M{}

	direction := c.DefaultQuery("direction", "7")

	directionInt, err := strconv.Atoi(direction)
	if err == nil && (directionInt == 1 || directionInt == 2) {
		filter["direction"] = directionInt
	}

	address := c.DefaultQuery("address", "")
	if address != "" && utils.IsValidEthereumAddress(address) {
		filter["$or"] = []bson.M{
			{"poolAddress": address},
		}
	}

	// 时间段
	recentdays := c.DefaultQuery("recentdays", "7")
	recentdaysInt, err := strconv.Atoi(recentdays)
	if err == nil {
		if recentdaysInt < 1 {
			recentdaysInt = 1
		} else if recentdaysInt > 180 {
			recentdaysInt = 180
		}

		date := time.Now().Add(-time.Duration(recentdaysInt) * time.Hour * 24)
		filter["eventTime"] = bson.M{
			"$gte": date,
		}

	}

	return &filter
}
