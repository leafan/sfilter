package controllers

import (
	"sfilter/api/utils"
	"sfilter/user/config"
	"sfilter/user/models"
	gutils "sfilter/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func SendCode(c *gin.Context) {
	address := c.DefaultQuery("username", "")
	if address == "" || !isValidEmail(address) {
		ResFailure(c, 401, "wrong email address")
		return
	}

	ip := c.ClientIP()

	// 检查下该 ip 最近 1小时有多少个验证码了, 如果超过一定数量, 进黑名单
	recent1hour := time.Now().Add(-60 * time.Minute)
	vcodes, _ := models.GetVerifyCodesByip(ip, recent1hour)
	if len(vcodes) > config.VerifyCodeMaxNum {
		// 一个小时之内某ip只允许3次发验证码机会
		ResFailure(c, 401, "Too many requests for verifycode, please request after 1 hour")
		return
	}

	// 先检查是否存在有效的code, 如果存在, 发回去
	recent10min := time.Now().Add(-10 * time.Minute) // 最近10min
	vcode, err := models.GetVerifyCodeByUser(address, recent10min)
	if err != nil || vcode == nil {
		// 没找到, 创建一个并发送邮件
		vcode = &models.VerifyCode{
			Username: address,
			ClientIp: ip,
			Code:     gutils.GenerateVerifyCode(6),
		}

		// 发送邮件
		// errMail := services.SendVerifyEmail(address, vcode.Code)
		// if errMail != nil {
		if false {
			ResFailure(c, 401, "send mail wrong, please contact deepeye support")
			return
		}

		// 邮件发送成功, 创建数据到db
		err = models.CreatVerifyCode(vcode.Username, vcode.Code, ip)
		if err != nil {
			ResFailure(c, 401, "send mail wrong, please contact deepeye support")
			return
		}
	}

	// 要么找到code了, 要么发送邮件了
	data := struct {
		Code string `json:"code"`
	}{
		Code: vcode.Code,
	}

	utils.ResSuccess(c, data)
}
