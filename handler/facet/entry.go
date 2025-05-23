package facet

import (
	"bytes"
	"encoding/hex"
	"sfilter/schema"

	"go.mongodb.org/mongo-driver/mongo"
)

const inscriptionDef = "data:,"
const facetApplicationDef = "data:application/vnd.facet.tx+json;rule=esip6,"

func HandleFacetLogic(block *schema.Block, mongodb *mongo.Client) {
	for _, tx := range block.Transactions {
		rawData := tx.OriginTx.Data()

		hexPrefix := []byte{0x64, 0x61, 0x74, 0x61, 0x3a} // "data:"
		if len(rawData) < len(hexPrefix) || !bytes.Equal(rawData[:len(hexPrefix)], hexPrefix) {
			// utils.Tracef("not inscription data: %x", rawData)
			continue
		}

		// 到这里, 说明是有 data: 开头了
		data, err := hex.DecodeString(hex.EncodeToString(rawData))
		if err != nil {
			// utils.Errorf("[ HandleFacetLogic ]  decode data error: %v", err)
			continue
		}
		// utils.Tracef("data: %v", string(data))

		// facet 逻辑
		if len(data) >= len(facetApplicationDef) && string(data[:len(facetApplicationDef)]) == facetApplicationDef {
			HandleFacet(block, tx, string(data[len(facetApplicationDef):]), mongodb)
		} else if len(data) >= len(inscriptionDef) && string(data[:len(inscriptionDef)]) == inscriptionDef {
			HandleInscription(block, tx, string(data[len(inscriptionDef):]), mongodb)
		}

	}
}
