package admin

import (
	"sfilter/api/utils"
	"sfilter/user/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AdminGetAllUsers(c *gin.Context) {
	options, err := parseAdminAllUsersOptions(c)
	if err != nil {
		return
	}

	filter := parseAdminAllUsersFilter(c)

	info, count, err := models.AdminGetAllUsersWithOptionFilter(options, filter)
	if err != nil {
		utils.ResFailure(c, 500, "Get data failed: "+err.Error())
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

func AdminUpdateRole(c *gin.Context) {
	var update models.AdminUpdateUserRoleForm
	err := c.ShouldBind(&update)
	if err != nil {
		utils.ResFailure(c, 401, "Wrong Params.")
		return
	}

	if !models.IsValidRole(update.Role) {
		// gutils.Tracef("wrong valid role, role: %v", update.Role)
		utils.ResFailure(c, 401, "Wrong role param")
		return
	}

	// modify role
	err = models.ResetUserRole(update.Username, update.Role)
	if err != nil {
		utils.ResFailure(c, 401, err.Error())
		return
	}

	utils.ResSuccess(c, "update role success.")
}

func parseAdminAllUsersOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := utils.ParsePageLimitParams(c)
	if err != nil {
		utils.ResFailure(c, 400, err.Error())
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
