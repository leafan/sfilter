package facet

import (
	"encoding/json"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/facet"
	"sfilter/utils"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleFacet(blk *schema.Block, tx *schema.Transaction, data string, mongodb *mongo.Client) {
	jsonData := []byte(data)
	var jsonObj schema.FacetJson

	err := json.Unmarshal(jsonData, &jsonObj)
	if err != nil {
		// utils.Errorf("[ HandleFacet ] unmarshal error: %v, data: %v", err, string(jsonData))
		return
	}
	// utils.Infof("[ HandleFacet ] debug.. json: %v", jsonObj)

	doHandleFacet(blk, tx, &jsonObj, string(jsonData), mongodb)
}

func doHandleFacet(blk *schema.Block, tx *schema.Transaction, jsn *schema.FacetJson, originData string, mongodb *mongo.Client) {
	// 先判断合法性
	userLogic := schema.UserLogic{
		Data:     originData,
		Op:       jsn.Op,
		Function: jsn.Data.Function,
		To:       jsn.Data.To,
	}

	val, ok := jsn.Data.Args["to"]
	if !ok {
		return
	}
	userLogic.ArgsTo = val.(string)

	if userLogic.Function == "swapExactTokensForTokens" {
		if amountIn, ok := jsn.Data.Args["amountIn"].(string); ok {
			amountInBig, ok := big.NewFloat(0).SetString(amountIn)
			if !ok {
				utils.Errorf("[ doHandleFacet ] set amountIn to float err, amountIn: %v", amountIn)
				return
			}

			userLogic.ArgsAmount, _ = amountInBig.Float64()
		}
	} else if userLogic.Function == "transfer" {
		if amountIn, ok := jsn.Data.Args["amountIn"].(string); ok {
			amountInBig, ok := big.NewFloat(0).SetString(amountIn)
			if !ok {
				utils.Errorf("[ doHandleFacet ] set amount to float err, amount: %v", amountIn)
				return
			}

			userLogic.ArgsAmount, _ = amountInBig.Float64()
		}
	}

	fct := schema.FacetModel{
		FacetProjectInfo: schema.FacetProjectInfo{
			ProjectAddress: tx.OriginTx.To().String(),
		},
		TxInfo: schema.TxInfo{
			BlockNo:  blk.BlockNo,
			TxHash:   tx.OriginTx.Hash().String(),
			GasPrice: tx.OriginTx.GasPrice().String(),
		},
		UserLogic: userLogic,

		CreatedAt: time.Unix(int64(blk.Block.Time()), 0),
	}

	fct.ProjectName = config.GetFacetProjectName(fct.ProjectAddress)

	sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
	if err != nil {
		utils.Warnf("[ doHandleFacet ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
	} else {
		fct.Operator = sender.String()
	}

	// utils.Infof("[ doHandleFacet ] fct: %v", fct)
	facet.SaveFacet(&fct, mongodb)
}
