package eth

import (
	"sfilter/api/utils"
	"sfilter/schema"
	"sfilter/services/global"
	"sfilter/services/pair"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetGlobalInfo(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	global, err := global.GetGlobalInfo(db)
	if err != nil {
		utils.ResFailure(c, 401, err.Error())
		return
	}

	hot, err := getHotPairs(db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	new, err := getNewHotPairs(db)
	if err != nil {
		utils.ResFailure(c, 500, err.Error())
		return
	}

	data := struct {
		Global *schema.GlobalInfo `json:"global"`
		Hot    []*schema.Pair     `json:"hot"`
		New    []*schema.Pair     `json:"new"`
	}{
		Global: global,
		Hot:    hot,
		New:    new,
	}

	utils.ResSuccess(c, data)
}

func getHotPairs(db *mongo.Database) ([]*schema.Pair, error) {
	filter := bson.M{}
	date := time.Now().Add(-1 * time.Hour)
	filter["updatedAt"] = bson.M{
		"$gte": date,
	}

	options := options.Find().SetSort(bson.D{{Key: "txNumIn1h", Value: -1}}).SetLimit(5)

	filter["$and"] = []bson.M{
		{"address": bson.M{"$ne": "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"}},
		{"address": bson.M{"$ne": "0x11b815efB8f581194ae79006d24E0d814B7697F6"}},
		{"address": bson.M{"$ne": "0x0d4a11d5EEaaC28EC3F61d100daF4d40471f1852"}},
		{"address": bson.M{"$ne": "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc"}},
		{"address": bson.M{"$ne": "0xc7bBeC68d12a0d1830360F8Ec58fA599bA1b0e9b"}},
	}

	hot, _, err := pair.GetHotPairs(options, &filter, db)
	if err != nil {
		return nil, err
	}

	return hot, nil
}

func getNewHotPairs(db *mongo.Database) ([]*schema.Pair, error) {
	filter := bson.M{}
	date := time.Now().Add(-time.Duration(24*7) * time.Hour) // 最近7天的新币
	filter["firstAddPoolTime"] = bson.M{
		"$gte": date,
	}
	dateUpdatedAt := time.Now().Add(-24 * time.Hour)
	filter["updatedAt"] = bson.M{
		"$gte": dateUpdatedAt,
	}

	options := options.Find().SetSort(bson.D{{Key: "txNumIn24h", Value: -1}}).SetLimit(5)

	hot, _, err := pair.GetHotPairs(options, &filter, db)
	if err != nil {
		return nil, err
	}

	return hot, nil
}
