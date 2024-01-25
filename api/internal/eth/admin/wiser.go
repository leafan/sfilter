package admin

import (
	"sfilter/api/utils"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/wiser"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AdminGetWisers(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	options, err := parseWiserOptions(c)
	if err != nil {
		return
	}

	filter := parseWiserFilter(c, db)

	info, count, err := wiser.GetWisers(options, filter, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	data := struct {
		List  []schema.Wiser `json:"list"`
		Count int64          `json:"count"`
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

func parseWiserOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}

	// 如果要查询地址, 以epoch排序, 否则以weight排序
	address := c.DefaultQuery("address", "")
	if address != "" {
		options = options.SetSort(bson.D{{Key: "epoch", Value: -1}})
	} else {
		options = options.SetSort(bson.D{{Key: "weight", Value: -1}})
	}

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")
	if sortBy != "" {
		key = sortBy
		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}

		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseWiserFilter(c *gin.Context, db *mongo.Database) *primitive.M {
	filter := bson.M{}

	address := c.DefaultQuery("address", "")
	if address != "" && utils.IsValidEthereumAddress(address) {
		filter["address"] = address
	}

	wiserType := c.DefaultQuery("type", "0")
	_type, err := strconv.Atoi(wiserType)
	if err == nil && (_type > 0 && _type < 10) {
		filter["type"] = _type
	}

	buyMevRatio := c.DefaultQuery("mevBuyRatio", "")
	if buyMevRatio == "1" {
		filter["buyMevRatio"] = bson.M{
			"$eq": 0,
		}
	}

	buyFreshRatio := c.DefaultQuery("freshBuyRatio", "")
	if buyFreshRatio == "1" {
		filter["buyFreshRatio"] = bson.M{
			"$eq": 0,
		}
	}

	// 默认直接传入最新的epoch, 但如果有过滤地址, 则不过滤epoch
	curEpoch, err := wiser.GetCurrentEpoch(db)
	if err == nil && address == "" {
		filter["epoch"] = curEpoch
	}

	// 前端传入的epoch, 则以前端为准
	epoch := c.DefaultQuery("epoch", "")
	if epoch != "" {
		filter["epoch"] = epoch
	}

	return &filter
}
