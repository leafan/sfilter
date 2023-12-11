package eth

import (
	"sfilter/api/utils"
	"sfilter/schema"
	"sfilter/services/global"
	"sfilter/services/pair"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var limit = int64(5)

func GetGlobalInfo(c *gin.Context) {
	global, err := global.GetGlobalInfo(utils.GetMongo())
	if err != nil {
		utils.ResFailure(c, 401, err.Error())
		return
	}

	hot, err := getHotPairs()
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	new, err := getNewHotPairs()
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	data := struct {
		Global *schema.GlobalInfo `json:"global"`
		Hot    []schema.Pair      `json:"hot"`
		New    []schema.Pair      `json:"new"`
	}{
		Global: global,
		Hot:    hot,
		New:    new,
	}

	utils.ResSuccess(c, data)
}

func getHotPairs() ([]schema.Pair, error) {
	filter := bson.M{}
	options := options.Find().SetSort(bson.D{{Key: "txNumIn24h", Value: -1}}).SetLimit(5)

	hot, _, err := pair.GetHotPairs(options, &filter, utils.GetMongo())
	if err != nil {
		return nil, err
	}

	return hot, nil
}

func getNewHotPairs() ([]schema.Pair, error) {
	filter := bson.M{}
	date := time.Now().Add(-time.Duration(24*7) * time.Hour) // 最近7天的新币
	filter["firstAddPoolTime"] = bson.M{
		"$gte": date,
	}

	options := options.Find().SetSort(bson.D{{Key: "txNumIn24h", Value: -1}}).SetLimit(5)

	hot, _, err := pair.GetHotPairs(options, &filter, utils.GetMongo())
	if err != nil {
		return nil, err
	}

	return hot, nil
}
