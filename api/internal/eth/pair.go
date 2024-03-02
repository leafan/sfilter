package eth

import (
	"sfilter/api/utils"
	"sfilter/schema"
	"sfilter/services/pair"
	services_utils "sfilter/utils"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetPair(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	address := c.DefaultQuery("address", "")
	if !utils.IsValidEthereumAddress(address) {
		utils.ResFailure(c, 500, "invalid params")
		return
	}

	data, err := pair.GetPairInfoForApi(address, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	var info []*schema.Pair
	info = append(info, data)
	// updateWrongData(info)

	utils.ResSuccess(c, info[0])
}

// hot pair(token): 最近n天 24h tx 数排名
// hot new pair(token): 最近7天添加的新池子中, 24h tx 数排名
func GetHotPairs(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	filter := parsePairFilterOptions(c)

	options, err := parsePairOptions(c, filter)
	if err != nil {
		return
	}

	info, count, err := pair.GetHotPairs(options, filter, db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	// 处理info的一些错误字段等, 如 txNumIn1h 可能由于1小时内都没有交易
	// 他也没有更新, 从而显示的还是老数据
	// updateWrongData(info)

	data := struct {
		List  []*schema.Pair `json:"list"`
		Count int64          `json:"count"`
	}{
		List:  info,
		Count: count,
	}

	utils.ResSuccess(c, data)
}

// 要根据排序条件, 增加查询条件
func parsePairOptions(c *gin.Context, filter *primitive.M) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}

	sort := bson.D{}

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")

	sorters := []string{"price", "txNumIn1h", "txNumIn24h", "txNumChangeIn1h", "txNumChangeIn24h", "priceChangeIn1h", "priceChangeIn24h", "volumeByUsdIn1h", "volumeByUsdIn24h", "firstAddPoolTime"}
	if services_utils.CheckExistString(sortBy, sorters) {
		key = sortBy

		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}

		(*filter)[key] = bson.M{
			"$ne": nil,
		}

		sort = append(sort, bson.E{Key: key, Value: order})

		// 如果是根据时间排序, 由于db是由swap触发更新, 如果这段时间没有更新, 则数据是老的
		// 因此如果是根据1h/24h等排序，则只过滤出近期数据有更新的data
		if key == "txNumIn1h" || key == "txNumChangeIn1h" || key == "priceChangeIn1h" || key == "volumeByUsdIn1h" {
			date := time.Now().Add(-1 * time.Hour)
			(*filter)["updatedAt"] = bson.M{
				"$gte": date,
			}
		}

		if key == "txNumIn24h" || key == "txNumChangeIn24h" || key == "priceChangeIn24h" || key == "volumeByUsdIn24h" {
			date := time.Now().Add(-24 * time.Hour)
			(*filter)["updatedAt"] = bson.M{
				"$gte": date,
			}
		}
	}

	sort = append(sort, bson.E{Key: "txNumIn1h", Value: -1}, bson.E{Key: "UpdateAt", Value: -1})
	options = options.SetSort(sort)

	return options, nil
}

func parsePairFilterOptions(c *gin.Context) *primitive.M {
	filter := bson.M{}

	// 根据token地址查询, 可能为token0或token1
	address := c.DefaultQuery("address", "")
	if address != "" && utils.IsValidEthereumAddress(address) {
		filter["$or"] = []bson.M{
			{"token0": address},
			{"token1": address},
			{"address": address},
		}
	}

	firstAddPoolHour := c.DefaultQuery("firstAddPoolHour", "")
	firstAddPoolHourInt, err := strconv.Atoi(firstAddPoolHour)
	if err == nil || firstAddPoolHourInt > 0 {
		if firstAddPoolHourInt > 180*60 {
			firstAddPoolHourInt = 180 * 60
		}

		date := time.Now().Add(-time.Duration(firstAddPoolHourInt) * time.Hour)
		filter["firstAddPoolTime"] = bson.M{
			"$gte": date,
		}
	}

	// 把常见的几个pair排除掉
	filter["$and"] = []bson.M{
		{"address": bson.M{"$ne": "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"}},
		{"address": bson.M{"$ne": "0x11b815efB8f581194ae79006d24E0d814B7697F6"}},
		{"address": bson.M{"$ne": "0x0d4a11d5EEaaC28EC3F61d100daF4d40471f1852"}},
		{"address": bson.M{"$ne": "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc"}},
		{"address": bson.M{"$ne": "0xc7bBeC68d12a0d1830360F8Ec58fA599bA1b0e9b"}},
	}

	return &filter
}
