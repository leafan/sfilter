package chain

import (
	"context"
	"log"
	"math/big"
	"sfilter/config"
	"sfilter/schema"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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

			token0, err0 := getSingleProp(address, "token0", getClient(), nil)
			token1, err1 := getSingleProp(address, "token1", getClient(), nil)

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

// 获取eth价格,
// 如果配置了height, 需要client支持archive查询功能, 可以用infura
func GetEthPrice(client *ethclient.Client, height *big.Int) float64 {
	if height != nil {
		client = getInfuraClient() // 当指定高度时, 则需要去infura上获取
	}

	const ETH_UNI_POOL = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"
	priceSqrt, err := getSingleProp(ETH_UNI_POOL, "slot0", client, height)

	if err != nil {
		log.Printf("[ GetEthPrice ] getSingleProp error: %v, height: %v\n", err, height)
		return 0
	}

	price := priceSqrt.(*big.Int)
	price = price.Mul(price, price)
	price = price.Mul(price, big.NewInt(1e18))

	newPrice := new(big.Float).SetInt(price)
	newPrice = newPrice.Quo(newPrice, new(big.Float).SetFloat64(1<<192))

	// 还需要处理decimal问题, 这里由于是固定eth价格且固定交易对, 直接写死了
	newPrice = newPrice.Quo(newPrice, new(big.Float).SetFloat64(1e30))
	newPrice = newPrice.Quo(new(big.Float).SetFloat64(1), newPrice)

	ret, _ := newPrice.Float64()

	ret = float64(int(ret*100)) / 100

	log.Printf("[ GetEthPrice ] block height: %v, price: %v\n\n", height, ret)
	return ret
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
	token0, _ := getSingleProp("0x42D52847BE255eacEE8c3f96b3B223c0B3cC0438", "token0", getClient(), nil) // v2
	token1, _ := getSingleProp("0xeA05D862E4c5CD0d3e660e0FCB2045C8DD4d7912", "token1", getClient(), nil) // v3

	log.Printf("\n\n[ TEST ] name: %v, symbol: %v\n\n\n", token0, token1)

	height := big.NewInt(18562231)
	ethPrice := GetEthPrice(getInfuraClient(), height)

	log.Printf("[ TEST ] eth price is: %v at height: %v\n\n\n", ethPrice, height)
}
