/**
 * @brief copy from: https://github.com/uture-ai/geg/blob/main/internal/util/cors.go
 */

package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//可将将*替换为指定的域名
			c.Header("Access-Control-Allow-Origin", "*")

			c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE,UPDATE")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}
