package encrypt

import (
	"sfilter/api/utils"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/transfer"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTransferEvents(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	options, err := parseTransferOptions(c)
	if err != nil {
		return
	}

	filter := parseTransferFilterOptions(c)

	info, count, err := transfer.GetTransferEvents(options, filter, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}
	data := struct {
		List  []schema.Transfer `json:"list"`
		Count int64             `json:"count"`
	}{
		List:  info,
		Count: count,
	}

	enc, err := utils.AesEncrypt(data, config.API_AES_DATA_KEY)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}
	utils.ResSuccess(c, enc)
}

func parseTransferOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "timestamp", Value: -1}})

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")
	if sortBy != "" && sortBy == "amount" {
		key = sortBy

		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}
		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseTransferFilterOptions(c *gin.Context) *primitive.M {
	filter := bson.M{}

	token := c.DefaultQuery("_token", "")
	if token != "" && utils.IsValidEthereumAddress(token) {
		filter["token"] = token
	}

	from := c.DefaultQuery("from", "")
	if from != "" && utils.IsValidEthereumAddress(from) {
		filter["from"] = from
	}

	to := c.DefaultQuery("to", "")
	if to != "" && utils.IsValidEthereumAddress(to) {
		filter["to"] = to
	}

	// 默认只查询 正常transfer 交易, 排除掉transfer交易
	filter["transferType"] = 1

	recentdays := c.DefaultQuery("recentdays", "1")
	recentdaysInt, err := strconv.Atoi(recentdays)
	if err == nil {
		if recentdaysInt < 1 {
			recentdaysInt = 1
		} else if recentdaysInt > 30 {
			recentdaysInt = 30
		}

		date := time.Now().Add(-time.Duration(recentdaysInt) * time.Hour * 24)
		filter["timestamp"] = bson.M{
			"$gte": date,
		}
	}

	return &filter
}
