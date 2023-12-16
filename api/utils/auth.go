package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthNothingMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端cookie并校验
		if true {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, "auth check fail: token is error, please check Authorization on your header")
		}

		c.Abort()
	}
}

func IPWhiteList(whitelist map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !whitelist[c.RemoteIP()] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "Permission denied",
			})

			log.Printf("[ IPWhiteList ] middleware failed.. ip: %v\n", c.RemoteIP())
			return
		}
	}
}
