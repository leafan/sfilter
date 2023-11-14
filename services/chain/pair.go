package chain

import (
	"context"
	"fmt"
	"log"
	"sfilter/config"
	"sfilter/schema"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPairInfo(address string) (*schema.Pair, error) {
	collection := getMongo().Database(config.DatabaseName).Collection(config.PairTableName)

	filter := bson.M{"address": address}

	var result schema.Pair
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 不存在，去链上查并返回
			result.Address = address

			token0, err0 := getSingleProp(address, "token0", getClient(), nil)
			token1, err1 := getSingleProp(address, "token1", getClient(), nil)

			if err0 != nil || err1 != nil {
				log.Printf("[ GetPairInfo ] get chainInfo error. address: %v, err0: %v, err1: %v\n", address, err0, err1)
				return nil, err1
			}
			result.Token0 = token0.(common.Address).String()
			result.Token1 = token1.(common.Address).String()

			// 存入mongo, 不判断错误, 只打印
			savePairInfo(&result, getMongo())

		} else {
			log.Printf("[ GetPairInfo ] FindOne error: %v, pair addr: %v\n", err, address)
			return nil, nil
		}

	} else {
		// log.Printf("[ GetPairInfo ] FindOne success. pair: %v, token0: %v\n", address, result.Token0)
	}

	return &result, nil
}

func savePairInfo(pair *schema.Pair, mongodb *mongo.Client) {
	log.Printf("[ savePairInfo ] InsertOne now. pair: %v\n", pair.Address)

	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	_, err := collection.InsertOne(context.Background(), pair)
	if err != nil {
		// 冲突挺多，不打印了
		// log.Printf("[ savePairInfo ] InsertOne error: %v, token: %v\n", err, pair.Address)
		return
	}

}

func TEST_PAIR() {
	pair1, _ := GetPairInfo("0x1901733a0b47eF6B4039D8b6451807660A5C85e4")
	pair2, _ := GetPairInfo("0x8802345e6b2b87fFa0290F799C84d00c6Eac5bb9")

	fmt.Printf("\n\n[ TEST ] pair1: %v, pair2: %v\n\n\n", pair1, pair2)

}
