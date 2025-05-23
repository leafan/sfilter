package admin

import (
	"sfilter/api/utils"
	"sfilter/config"

	wiser "sfilter/handler/wiser"

	"github.com/gin-gonic/gin"
)

func GetHotBigPairs(c *gin.Context) {
	set := wiser.NewSetting("", config.DatabaseName, false)
	hndl := &wiser.Handler{
		Hbpair: &wiser.HBPair{
			Set: set,
		},
	}

	h1, h4, h24 := hndl.Hbpair.HotPairSearch()

	data := struct {
		H1  []string `json:"h1"`
		H4  []string `json:"h4"`
		H24 []string `json:"h24"`
	}{
		H1:  h1,
		H4:  h4,
		H24: h24,
	}

	utils.ResSuccess(c, data)
}
