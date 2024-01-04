package facet

import (
	"encoding/json"
	"sfilter/schema"
	"sfilter/services/facet"
	"sfilter/utils"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// 普通铭文
func HandleInscription(blk *schema.Block, tx *schema.Transaction, data string, mongodb *mongo.Client) {
	jsonData := []byte(data)
	var jsonObj schema.Inscription

	err := json.Unmarshal(jsonData, &jsonObj)
	if err != nil {
		// 说明不是标准铭文
		// utils.Warnf("[ HandleInscription ] unmarshal error: %v, data: %v", err, string(jsonData))
		doHandleInscription(blk, tx, nil, string(jsonData), mongodb)
	} else {
		doHandleInscription(blk, tx, &jsonObj, string(jsonData), mongodb)
	}
}

func doHandleInscription(blk *schema.Block, tx *schema.Transaction, jsn *schema.Inscription, originData string, mongodb *mongo.Client) {
	ins := schema.InscriptionModel{
		Data: originData,
		TxInfo: schema.TxInfo{
			BlockNo:  blk.BlockNo,
			TxHash:   tx.OriginTx.Hash().String(),
			GasPrice: tx.OriginTx.GasPrice().String(),
		},

		CreatedAt: time.Unix(int64(blk.Block.Time()), 0),
	}
	if jsn != nil {
		ins.Inscription = *jsn
	}

	sender, err := types.Sender(types.NewLondonSigner(tx.OriginTx.ChainId()), tx.OriginTx)
	if err != nil {
		utils.Warnf("[ doHandleInscription ] types.Sender err: %v, tx: %v\n", err, tx.OriginTx.Hash())
	} else {
		ins.Operator = sender.String()
	}

	facet.SaveInscription(&ins, mongodb)
}
