package admin

import (
	"sfilter/api/utils"
	"sfilter/user/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AdminGetAllTrackedAddress(c *gin.Context) {
	options, err := parseAdminAllTrackAddrsOptions(c)
	if err != nil {
		return
	}

	filter := parseAdminAllTrackAddrsFilter(c)

	info, count, err := models.AdminGetAllTrackAddrsWithOptionFilter(options, filter)
	if err != nil {
		utils.ResFailure(c, 500, "Get data failed: "+err.Error())
		return
	}

	data := struct {
		List  []models.UserTrackedAddress `json:"list"`
		Count int64                       `json:"count"`
	}{
		List:  info,
		Count: count,
	}

	utils.ResSuccess(c, data)
}

func parseAdminAllTrackAddrsOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	return options, nil
}

func parseAdminAllTrackAddrsFilter(c *gin.Context) *primitive.M {
	filter := bson.M{}

	finds := c.DefaultQuery("finds", "")
	if finds != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"username": bson.M{"$regex": finds, "$options": "i"}},
				{"address": bson.M{"$regex": finds, "$options": "i"}},
			},
		}
	}

	return &filter
}
