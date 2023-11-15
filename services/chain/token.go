package chain

import (
	"context"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	service_token "sfilter/services/token"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTokenInfo(address string) (*schema.Token, error) {
	collection := getMongo().Database(config.DatabaseName).Collection(config.TokenTableName)

	filter := bson.D{{Key: "address", Value: address}}

	var result schema.Token
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 不存在，去链上查并返回
			result.Address = address

			var err1, err2 error
			var decimals, symbol interface{}
			decimals, err1 = getSingleProp(address, "decimals", getClient(), nil)
			symbol, err2 = getSingleProp(address, "symbol", getClient(), nil)

			// name和totalsupply没取到也没事
			name, err3 := getSingleProp(address, "name", getClient(), nil)
			totalsupply, err4 := getSingleProp(address, "totalSupply", getClient(), nil)

			if err1 != nil || err2 != nil {
				log.Printf("[ GetTokenInfo ] get chainInfo error. address: %v, err1: %v, err2: %v\n", address, err1, err2)
				return nil, err2
			}
			result.Decimal = decimals.(uint8)
			result.Name = name.(string)

			if err3 != nil || err4 != nil {
				log.Printf("[ GetTokenInfo ] get chainInfo error. address: %v, err3: %v, err4: %v\n", address, err3, err4)
				// 不退出
			} else {
				result.Symbol = symbol.(string)
				result.TotalSupply = totalsupply.(*big.Int).String()
			}

			service_token.SaveTokenInfo(&result, getMongo())
		} else {
			log.Printf("[ GetTokenInfo ] FindOne error: %v, token addr: %v\n", err, address)
			return nil, err
		}

	} else {
		// log.Printf("[ GetTokenInfo ] FindOne success, token addr: %v\n", address)
	}

	return &result, nil
}

func TEST_TOKEN() {
	token, _ := GetTokenInfo("0x180EFC1349A69390aDE25667487a826164C9c6E4")

	log.Printf("\n\n[ TEST ] token: %v,\n\n\n", token)

	service_token.UpdateTokenInfo(token, getMongo())
}
