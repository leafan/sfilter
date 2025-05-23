package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code uint32      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ResSuccess(c *gin.Context, data interface{}) {
	result := Response{
		Code: http.StatusOK,
		Msg:  "SUCCESS",
		Data: data,
	}
	c.JSON(http.StatusOK, result)
}

func ResFailure(c *gin.Context, code uint32, message string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  message,
		Data: nil,
	})

	c.Abort()
}
