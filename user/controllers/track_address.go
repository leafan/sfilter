package controllers

import (
	"sfilter/api/utils"
	"sfilter/user/models"
	gutils "sfilter/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-pkgz/auth/token"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTrackedAddresses(c *gin.Context) {
	user, err := token.GetUserInfo(c.Request)
	if err != nil {
		ResFailure(c, 401, "Wrong Claims.")
		return
	}

	options, err := parseTrackAddressOptions(c)
	if err != nil {
		return
	}

	filter := parseTrackedAddressFilter(c)

	info, count, err := models.GetTrackedAddressesWithOptionFilter(user.Name, options, filter)
	if err != nil {
		ResFailure(c, 500, "Get data failed.")
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

func CreateTrackedAddress(c *gin.Context) {
	user, err := token.GetUserInfo(c.Request)
	if err != nil {
		ResFailure(c, 401, "Wrong Claims.")
		return
	}

	var addr models.TrackAddressForm
	err = c.ShouldBind(&addr)
	if err != nil {
		gutils.Tracef("err: %v, addr: %v", err, addr)
		ResFailure(c, 403, "Wrong Params.")
		return
	}

	if !utils.IsValidEthereumAddress(addr.Address) {
		utils.ResFailure(c, 500, "Invalid address")
		return
	}

	if len(addr.Memo) > 30 {
		utils.ResFailure(c, 403, "Memo is too long.")
		return
	}

	_, err = models.GetEntryByUserAndAddress(user.Name, addr.Address)
	if err == nil {
		ResFailure(c, 401, "Address exists.")
		return
	}

	_, err = models.GetEntryByUserAndMemo(user.Name, addr.Memo)
	if err == nil {
		ResFailure(c, 401, "Memo exists.")
		return
	}

	// 判断是否已经创建了太多地址, 如果是请升级
	count, err := models.GetUserTrackAddressCount(user.Name)
	if err != nil {
		ResFailure(c, 500, "System error to finish request")
		return
	}

	role, ok := user.Attributes["role"]
	if !ok {
		ResFailure(c, 500, "Wrong roles")
		return
	}

	if count > models.GetRoleTrackAddressCount(role.(string)) {
		ResFailure(c, 403, "Too many tracked address, please upgrade your level or contact deepeye team")
		return
	}

	err = models.CreateTrackedAddress(user.Name, addr.Address, addr.Memo, addr.Priority)
	if err != nil {
		ResFailure(c, 500, "System error to creat tracked address")
		return
	}

	ResSuccess(c, "Create success")
}

func UpdateTrackedAddress(c *gin.Context) {
	user, err := token.GetUserInfo(c.Request)
	if err != nil {
		ResFailure(c, 401, "Wrong Claims.")
		return
	}

	address := c.Param("address")
	if address == "" || !utils.IsValidEthereumAddress(address) {
		ResFailure(c, 401, "Wrong address params.")
		return
	}

	// 先检查是否存在
	_, err = models.GetEntryByUserAndAddress(user.Name, address)
	if err != nil {
		ResFailure(c, 401, "No such address")
		return
	}

	var update models.UpdateTrackAddressForm
	err = c.ShouldBind(&update)
	if err != nil {
		ResFailure(c, 401, "Wrong Params.")
		return
	}

	err = models.UpdateTrackedAddress(user.Name, address, update.Memo, update.Priority)
	if err != nil {
		ResFailure(c, 500, "System error to creat tracked address")
		return
	}

	ResSuccess(c, "Update success")
}

func DeleteTrackedAddress(c *gin.Context) {
	user, err := token.GetUserInfo(c.Request)
	if err != nil {
		ResFailure(c, 401, "Wrong Claims.")
		return
	}

	address := c.Param("address")
	if address == "" || !utils.IsValidEthereumAddress(address) {
		ResFailure(c, 401, "Wrong address params.")
		return
	}

	_, err = models.GetEntryByUserAndAddress(user.Name, address)
	if err != nil {
		gutils.Tracef("no such addr, user:  %v, addr:  %v, err: %v", user.Name, address, err)
		ResFailure(c, 401, "No such address")
		return
	}

	err = models.DeleteTrackedAddress(user.Name, address)
	if err != nil {
		ResFailure(c, 500, "System error to delete tracked address")
		return
	}

	ResSuccess(c, "Delete success")
}

func parseTrackAddressOptions(c *gin.Context) (*options.FindOptions, error) {
	page, limit, err := ParsePageLimitParams(c)
	if err != nil {
		ResFailure(c, 400, err.Error())
		return nil, err
	}

	skip := int64(page*limit - limit)
	options := &options.FindOptions{Limit: &limit, Skip: &skip}
	options = options.SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	// 获取排序
	var key string
	var order = -1
	sortBy := c.DefaultQuery("sortBy", "")
	if sortBy == "updatedAt" || sortBy == "priority" {
		key = sortBy
		orderStr := c.DefaultQuery("descending", "true")
		if orderStr == "false" {
			order = 1
		}

		options = options.SetSort(bson.D{{Key: key, Value: order}})
	}

	return options, nil
}

func parseTrackedAddressFilter(c *gin.Context) *primitive.M {
	filter := bson.M{}

	finds := c.DefaultQuery("finds", "")
	if finds != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"address": bson.M{"$regex": finds, "$options": "i"}},
				{"memo": bson.M{"$regex": finds, "$options": "i"}},
			},
		}
	}

	return &filter
}
