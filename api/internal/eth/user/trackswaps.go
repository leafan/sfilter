package user

import (
	"sfilter/api/utils"
	"sfilter/schema"
	"sfilter/services/swap"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTrackSwaps(c *gin.Context) {
	var username string
	user, err := token.GetUserInfo(c.Request)
	if err != nil {
		// 判断apikey
		username = c.GetString("user")
		if username == "" {
			utils.ResFailure(c, 401, "Wrong Claims/ApiKey.")
			return
		}
	} else {
		username = user.Name
	}

	db := utils.GetChainDatabase(c.Param("chain"))

	options, err := parseTrackSwapOptions(c)
	if err != nil {
		return
	}

	// 将用户名传入查询
	filter := parseTrackSwapFilterOptions(username, c)

	info, count, err := swap.GetTrackSwaps(options, filter, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	data := struct {
		List  []schema.TrackSwap `json:"list"`
		Count int64              `json:"count"`
	}{
		List:  info,
		Count: count,
	}

	utils.ResSuccess(c, data)
}

func parseTrackSwapOptions(c *gin.Context) (*options.FindOptions, error) {
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
	if sortBy != "" && (sortBy == "amountOfMainToken" || sortBy == "volumeInUsd" || sortBy == "priority") {
		key = sortBy

		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}
		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseTrackSwapFilterOptions(username string, c *gin.Context) *primitive.M {
	filter := bson.M{
		"username": username,
	}

	direction := c.DefaultQuery("direction", "0")
	directionInt, err := strconv.Atoi(direction)
	if err == nil && (directionInt == 1 || directionInt == 2) {
		filter["direction"] = directionInt
	}

	trader := c.DefaultQuery("trader", "")
	if trader != "" {
		orCondition := []bson.M{{"address": trader}, {"memo": trader}}
		filter["$or"] = append(filter["$or"].([]bson.M), orCondition...)
	}

	pairAddr := c.DefaultQuery("pairAddr", "")
	if pairAddr != "" && utils.IsValidEthereumAddress(pairAddr) {
		filter["pairAddr"] = pairAddr
	}

	// 根据token地址查询, 可能为token0或token1
	token := c.DefaultQuery("_token", "")
	if token != "" && utils.IsValidEthereumAddress(token) {
		orCondition := []bson.M{
			{"token0": token},
			{"token1": token},
			{"pairAddr": token},
		}
		filter["$or"] = append(filter["$or"].([]bson.M), orCondition...)
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
