package pair

import (
	"context"
	"errors"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPairInfoForRead(address string) (*schema.Pair, error) {
	mongodb := chain.GetMongo()
	return GetPairInfo(address, mongodb)
}

func GetPairInfo(address string, mongodb *mongo.Client) (*schema.Pair, error) {
	if address == "" {
		log.Printf("[ GetPairInfo ] error! empty address. address!")
		return nil, errors.New("empty address")
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	filter := bson.M{"address": address}

	var result schema.Pair
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 不存在，去链上查并返回
			result.Address = address

			token0, err0 := chain.GetSingleProp(address, "token0")
			token1, err1 := chain.GetSingleProp(address, "token1")

			if err0 != nil || err1 != nil {
				// log.Printf("[ GetPairInfo ] get chainInfo error. address: %v, err0: %v, err1: %v\n", address, err0, err1)
				return nil, err1
			}
			result.Token0 = token0.(common.Address).String()
			result.Token1 = token1.(common.Address).String()

			// 存入mongo, 不判断错误, 只打印
			UpSertOnChainInfo(result.Address, &result.InfoOnChain, mongodb)

		} else {
			log.Printf("[ GetPairInfo ] FindOne error: %v, pair addr: %v\n", err, address)
			return nil, nil
		}

	}

	return &result, nil
}

func TEST_PAIR() {
	// GetPairInfo("0x58Dc5a51fE44589BEb22E8CE67720B5BC5378009", getMongo())

	pairx := &schema.Pair{
		InfoOnChain: schema.InfoOnChain{
			Address:  "0x58Dc5a51fE44589BEb22E8CE67720B5BC5378009",
			PairName: "leafan6",
		},
	}

	UpSertOnChainInfo("0x58Dc5a51fE44589BEb22E8CE67720B5BC5378009", &pairx.InfoOnChain, chain.GetMongo())

}
