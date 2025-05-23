package eth

import (
	"sfilter/api/utils"
	"sfilter/schema"
	"sfilter/services/kline"
	"sfilter/services/pair"
	"time"

	"github.com/gin-gonic/gin"
)

// 获取 1h 数据(分钟线) 与 48h 数据(小时线)
// 返回他们的柱子
func GetPriceAndTxTrends(c *gin.Context) {
	db := utils.GetChainDatabase(c.Param("chain"))

	address := c.DefaultQuery("address", "")
	if !utils.IsValidEthereumAddress(address) {
		utils.ResFailure(c, 500, "invalid params")
		return
	}

	now := time.Now()
	klines1Min := kline.Get1MinKlineWithFullGenerated(address, now, 1, db)
	klines1Hour := kline.Get1HourKlineWithFullGenerated(address, now, 2, db)
	// fmt.Printf("klines1Hour: %v\n", klines1Hour)

	var p1, t1, p48, t48 schema.TrendStruct
	for _, v := range klines1Min {
		if v.UnixTime <= 0 {
			continue
		}

		elemP := schema.TrendElem{
			Value: v.ClosePrice,
			Time:  time.Unix(v.UnixTime, 0),
		}
		elemT := schema.TrendElem{
			Value: float64(v.TxNum),
			Time:  time.Unix(v.UnixTime, 0),
		}

		p1 = append(p1, elemP)
		t1 = append(t1, elemT)
	}

	for _, v := range klines1Hour {
		if v.UnixTime <= 0 {
			continue
		}

		elemP := schema.TrendElem{
			Value: v.ClosePrice,
			Time:  time.Unix(v.UnixTime, 0),
		}
		elemT := schema.TrendElem{
			Value: float64(v.TxNum),
			Time:  time.Unix(v.UnixTime, 0),
		}

		p48 = append(p48, elemP)
		t48 = append(t48, elemT)
	}

	price, err := pair.GetPairInfoForRead(address)
	if err != nil {
		utils.ResFailure(c, 500, "Wrong pair, can not get the price.")
		return
	}

	data := struct {
		Price     float64            `json:"price"`
		Prices1h  schema.TrendStruct `json:"prices1h"`
		Txs1h     schema.TrendStruct `json:"txs1h"`
		Prices48h schema.TrendStruct `json:"prices48h"`
		Txs48h    schema.TrendStruct `json:"txs48h"`
	}{
		Price:     price.Price,
		Prices1h:  p1,
		Txs1h:     t1,
		Prices48h: p48,
		Txs48h:    t48,
	}

	utils.ResSuccess(c, data)
}
