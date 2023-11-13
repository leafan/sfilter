package chain

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTokenInfo(address string, mongodb *mongo.Client) (*schema.Token, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	filter := bson.D{{Key: "address", Value: address}}

	var result schema.Token
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 不存在，去链上查并返回
			result.Address = address

			var err1, err2, err3 error
			var name, symbol interface{}
			result.Decimal, err1 = getTokenDecimals(address)
			name, err2 = getSingleProp(address, "name")
			symbol, err3 = getSingleProp(address, "symbol")

			if err1 != nil || err2 != nil || err3 != nil {
				log.Printf("[ GetTokenInfo ] get chainInfo error. address: %v, err1: %v, err2: %v, err3: %v\n", address, err1, err2, err3)
				return nil, err3
			}
			result.Name = name.(string)
			result.Symbol = symbol.(string)

			// 存入mongo, 不判断错误, 只打印
			saveTokenInfo(&result, mongodb)

		} else {
			log.Printf("[ GetTokenInfo ] FindOne error: %v, token addr: %v\n", err, address)
			return nil, err
		}

	}

	return &result, nil
}

func saveTokenInfo(token *schema.Token, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TokenTableName)

	_, err := collection.InsertOne(context.Background(), token)
	if err != nil {
		log.Printf("[ saveTokenInfo ] InsertOne error: %v, token: %v\n", err, token.Address)
		return
	}

}

func getSingleProp(address, info string) (interface{}, error) {
	abi := getAbi()
	client := getClient()
	contractAddr := common.HexToAddress(address)
	bytes, _ := abi.Pack(info)
	msg := ethereum.CallMsg{
		From: common.Address{},
		To:   &contractAddr,
		Data: bytes,
	}

	ret, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Printf("[ getTokenNameOrSymbol ] CallContract error. addr: %v, err: %v\n", address, err)
		return "", err
	}

	intr, err := abi.Methods[info].Outputs.UnpackValues(ret)
	if err != nil {
		log.Printf("[ getTokenNameOrSymbol ] UnpackValues error. addr: %v, err: %v\n", address, err)
		return "", err
	}

	return intr[0], err
}

func getTokenDecimals(address string) (uint8, error) {
	abi := getAbi()
	client := getClient()
	contractAddr := common.HexToAddress(address)
	bytes, _ := abi.Pack("decimals")
	msg := ethereum.CallMsg{
		From: common.Address{},
		To:   &contractAddr,
		Data: bytes,
	}

	// 每一笔CallContract时间消耗在 0.3-0.5ms 之间
	ret, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Printf("[ getTokenDecimals ] CallContract error. addr: %v, err: %v\n", address, err)
		return 0, err
	}

	intr, err := abi.Methods["decimals"].Outputs.UnpackValues(ret)
	if err != nil {
		log.Printf("[ getTokenDecimals ] UnpackValues error. addr: %v, err: %v\n", address, err)
		return 0, err
	}

	return intr[0].(uint8), nil
}

func TEST_TOKEN() {
	decimals, _ := getTokenDecimals("0x42D52847BE255eacEE8c3f96b3B223c0B3cC0438")
	log.Printf("\n\n[ TEST ] decimals: %v\n\n\n", decimals)

	name, _ := getSingleProp("0x42D52847BE255eacEE8c3f96b3B223c0B3cC0438", "name")
	symbol, _ := getSingleProp("0x42D52847BE255eacEE8c3f96b3B223c0B3cC0438", "symbol")
	log.Printf("\n\n[ TEST ] name: %v, symbol: %v\n\n\n", name, symbol)
}
