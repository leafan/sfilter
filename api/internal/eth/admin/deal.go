package admin

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

func AdminGetAllDeals(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	options, err := parseDealOptions(c)
	if err != nil {
		return
	}

	filter := parseDealFilter(c)

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

func parseDealOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")
	if sortBy == "buyAmount" || sortBy == "sellAmount" || sortBy == "holdBlocks" || sortBy == "earnChange" || sortBy == "earn" {
		key = sortBy
		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}

		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseDealFilter(c *gin.Context) *primitive.M {
	filter := bson.M{}

	account := c.DefaultQuery("account", "")
	if account != "" && utils.IsValidEthereumAddress(account) {
		filter["account"] = account
	}

	token := c.DefaultQuery("_token", "")
	if token != "" && utils.IsValidEthereumAddress(token) {
		filter["token"] = token
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
		filter["createdAt"] = bson.M{
			"$gte": date,
		}

	}

	return &filter
}
