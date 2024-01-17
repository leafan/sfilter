package utils

import (
	"errors"
	"sfilter/config"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParsePageLimitParams(c *gin.Context) (int64, int64, error) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("pageSize", "10")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return 0, 0, errors.New("invalid page parameter")
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return 0, 0, errors.New("invalid limit parameter")
	}

	if pageInt <= 0 {
		pageInt = 1
	}

	if pageInt > config.MONGO_PAGE_UPPER {
		pageInt = config.MONGO_PAGE_UPPER
	}

	if limitInt < config.MONGO_LIMIT_DOWN {
		limitInt = config.MONGO_LIMIT_DOWN
	}

	if limitInt > config.MONGO_LIMIT_UPPER {
		// 特殊处理下, 如果是apikey访问, 允许增大到1000条数据
		username := c.GetString("user")
		if username == "" || limitInt > config.MONGO_APIKEY_LIMIT_UPPER {
			limitInt = config.MONGO_LIMIT_UPPER
		}
	}

	return int64(pageInt), int64(limitInt), nil
}
