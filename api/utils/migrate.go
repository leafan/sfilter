package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

func EmptyMigrate(group *gin.RouterGroup) {
	group.POST("/migration", func(c *gin.Context) {
		go func() {
			log.Println("[ migrate ] post migration now.. return")
		}()

		ResSuccess(c, gin.H{"message": "no migration yet"})
	})
}
