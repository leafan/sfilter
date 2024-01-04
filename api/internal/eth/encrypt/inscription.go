package encrypt

import (
	"sfilter/api/utils"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/facet"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetInscriptions(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	options, err := parseInscriptionOptions(c)
	if err != nil {
		return
	}

	filter := parseInscriptionFilterOptions(c)

	info, count, err := facet.GetInscriptions(options, filter, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}
	data := struct {
		List  []schema.InscriptionModel `json:"list"`
		Count int64                     `json:"count"`
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

func parseInscriptionOptions(c *gin.Context) (*options.FindOptions, error) {
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
	if sortBy != "" && sortBy == "id" {
		key = sortBy

		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}
		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseInscriptionFilterOptions(c *gin.Context) *primitive.M {
	filter := bson.M{}

	operator := c.DefaultQuery("operator", "")
	if operator != "" && utils.IsValidEthereumAddress(operator) {
		filter["operator"] = operator
	}

	tick := c.DefaultQuery("tick", "")
	if tick != "" {
		filter["tick"] = tick
	}

	op := c.DefaultQuery("op", "")
	if op != "" {
		filter["op"] = op
	}

	p := c.DefaultQuery("p", "")
	if p != "" {
		filter["p"] = p
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
		filter["createdAt"] = bson.M{
			"$gte": date,
		}
	}

	return &filter
}
