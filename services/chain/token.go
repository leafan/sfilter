package chain

import (
	"bytes"
	"context"
	"errors"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"
	service_token "sfilter/services/token"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 主要用于Test，自己获取mongo玩
func GetTokenInfoForRead(address string) (*schema.Token, error) {
	mongodb := getMongo()
	return GetTokenInfo(address, mongodb)
}

func GetTokenInfo(address string, mongodb *mongo.Client) (*schema.Token, error) {
	if address == "" {
		log.Printf("[ GetTokenInfo ] error! address: %v\n", address)
		return nil, errors.New("address is empty")
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	filter := bson.D{{Key: "address", Value: address}}

	var result schema.Token
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 不存在，去链上查并返回
			result.Address = address

			decimals, err1 := getSingleProp(address, "decimals", getClient(), nil)
			if err1 != nil {
				log.Printf("[ GetTokenInfo ] get decimals error. set to 0 now. address: %v, err1: %v\n", address, err1)
				result.Decimal = 0
			} else {
				result.Decimal = decimals.(uint8)
			}

			// name和totalsupply没取到也没事
			name, err3 := getSingleProp(address, "name", getClient(), nil)
			if err3 != nil {
				name, err3 = getSingleBackupProp(address, "name", getClient(), nil)
				if err3 == nil {
					name2 := name.([32]byte)
					result.Name = string(bytes.TrimRight(name2[:], "\x00"))
				}
			} else {
				result.Name = name.(string)
			}

			symbol, err2 := getSingleProp(address, "symbol", getClient(), nil)
			if err2 != nil {
				// 再取一次, 还失败就不要了
				symbol, err21 := getSingleBackupProp(address, "symbol", getClient(), nil)
				if err21 != nil {
					// 如果name存在, 则把name赋值给symbol
					if result.Name != "" {
						result.Symbol = result.Name
					} else {
						// symbol必须取到, pair需要用
						return nil, err21
					}
				} else {
					symbol2 := symbol.([32]byte)
					result.Symbol = string(bytes.TrimRight(symbol2[:], "\x00"))
				}
			} else {
				result.Symbol = symbol.(string)
			}

			totalsupply, err4 := getSingleProp(address, "totalSupply", getClient(), nil)
			if err4 != nil {
				result.TotalSupply = big.NewInt(0).String()
			} else {
				result.TotalSupply = totalsupply.(*big.Int).String()
			}

			// save...
			service_token.SaveTokenInfo(&result, mongodb)
		}
	}

	return &result, nil
}

func TEST_TOKEN() {
	token, _ := GetTokenInfoForRead("0xC19B6A4Ac7C7Cc24459F08984Bbd09664af17bD1")

	log.Printf("\n\n[ TEST ] token: %v,\n\n\n", token)

	// service_token.UpdateTokenInfo(token, getMongo())

	token2, _ := GetTokenInfoForRead("0x431ad2ff6a9C365805eBaD47Ee021148d6f7DBe0")
	log.Printf("\n\n[ TEST ] token2: %v,\n\n\n", token2)

}
