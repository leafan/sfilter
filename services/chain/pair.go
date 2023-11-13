package chain

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPairInfo(address string, mongodb *mongo.Client) (*schema.Pair, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	filter := bson.D{{Key: "address", Value: address}}

	var result schema.Pair
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 不存在，去链上查并返回
			result.Address = address

			token0, err0 := getSingleProp(address, "token0")
			token1, err1 := getSingleProp(address, "token1")

			if err0 != nil || err1 != nil {
				log.Printf("[ GetPairInfo ] get chainInfo error. address: %v, err0: %v, err1: %v\n", address, err0, err1)
				return nil, err1
			}
			result.Token0 = token0.(common.Address).String()
			result.Token1 = token1.(common.Address).String()

			// 存入mongo, 不判断错误, 只打印
			savePairInfo(&result, mongodb)

		} else {
			log.Printf("[ GetPairInfo ] FindOne error: %v, pair addr: %v\n", err, address)
			return nil, err
		}

	}

	return &result, nil
}

func savePairInfo(pair *schema.Pair, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	_, err := collection.InsertOne(context.Background(), pair)
	if err != nil {
		log.Printf("[ savePairInfo ] InsertOne error: %v, token: %v\n", err, pair.Address)
		return
	}

}

func TEST_PAIR() {
	token0, _ := getSingleProp("0x42D52847BE255eacEE8c3f96b3B223c0B3cC0438", "token0") // v2
	token1, _ := getSingleProp("0xeA05D862E4c5CD0d3e660e0FCB2045C8DD4d7912", "token1") // v3

	log.Printf("\n\n[ TEST ] name: %v, symbol: %v\n\n\n", token0, token1)
}
