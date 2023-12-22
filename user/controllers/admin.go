package controllers

import (
	"fmt"
	"sfilter/api/utils"
	"sfilter/user/models"

	"github.com/gin-gonic/gin"
	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AdminGetAllUsers(c *gin.Context) {
	user, err := token.GetUserInfo(c.Request)
	if err != nil {
		ResFailure(c, 401, "Wrong Claims.")
		return
	}

	if user.Attributes["role"] != fmt.Sprintf("%d", models.USER_ROLE_LEVEL_ROOT) {
		ResFailure(c, 401, "No privilige to execute")
		return
	}

	options, err := parseAdminAllUsersOptions(c)
	if err != nil {
		return
	}

	filter := parseAdminAllUsersFilter(c)

	info, count, err := models.AdminGetAllUsersWithOptionFilter(options, filter)
	if err != nil {
		ResFailure(c, 500, "Get data failed.")
		return
	}

	data := struct {
		List  []models.User `json:"list"`
		Count int64         `json:"count"`
	}{
		List:  info,
		Count: count,
	}
	utils.ResSuccess(c, data)

}

func parseAdminAllUsersOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := ParsePageLimitParams(c)
	if err != nil {
		ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "registerAt", Value: -1}})

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")
	if sortBy == "registerAt" {
		key = sortBy
		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}

		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseAdminAllUsersFilter(c *gin.Context) *primitive.M {
	filter := bson.M{}

	finds := c.DefaultQuery("finds", "")
	if finds != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"username": bson.M{"$regex": finds, "$options": "i"}},
				{"email": bson.M{"$regex": finds, "$options": "i"}},
				{"parent": bson.M{"$regex": finds, "$options": "i"}},
			},
		}
	}

	return &filter
}
