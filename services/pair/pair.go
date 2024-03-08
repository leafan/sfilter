package pair

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/utils"

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
		return nil, errors.New("GetPairInfo: empty address")
	}

	collection := mongodb.Database(config.DatabaseName).Collection(config.PairTableName)

	filter := bson.M{"address": address}

	var result schema.Pair
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 不存在，去链上查并返回
			_pair := getPairOnChainInfoFromChain(address, mongodb)
			if _pair != nil {
				UpSertOnChainInfo(_pair.Address, &_pair.InfoOnChain, mongodb)
				return _pair, nil
			}

			return nil, errors.New("getPairOnChainInfoFromChain failed")
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func GetPairInfoForApi(address string, db *mongo.Database) (*schema.Pair, error) {
	collection := db.Collection(config.PairTableName)

	filter := bson.M{"address": address}

	var result schema.Pair
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// 当db中不存在时，update一下信息
func getPairOnChainInfoFromChain(address string, mongodb *mongo.Client) *schema.Pair {
	var pair schema.Pair

	pair.Address = address
	token0Addr, err0 := chain.GetSingleProp(address, "token0")
	token1Addr, err1 := chain.GetSingleProp(address, "token1")

	if err0 != nil || err1 != nil {
		return nil
	}
	pair.Token0 = token0Addr.(common.Address).String()
	pair.Token1 = token1Addr.(common.Address).String()

	token0, err0 := chain.GetTokenInfo(pair.Token0, mongodb)
	token1, err1 := chain.GetTokenInfo(pair.Token1, mongodb)
	if err0 != nil || err1 != nil {
		log.Printf("[ getPairInfoOnChain ] getTokenInfo error. token0: %v, err0: %v, token1: %v, err1: %v\n", pair.Token0, err0, pair.Token1, err1)
		return nil
	}

	pair.Decimal0 = token0.Decimal
	pair.Decimal1 = token1.Decimal

	GeneratePairName(&pair, token0, token1)

	return &pair
}

func GeneratePairName(pair *schema.Pair, token0, token1 *schema.Token) {
	// 组装pairName, 如果一方是价值token, 则作为quoteToken
	pair.PairName = fmt.Sprintf("%s/%s", token0.Symbol, token1.Symbol)

	quoteToken := utils.GetQuoteToken(pair.Token0, pair.Token1)
	if quoteToken == pair.Token0 {
		pair.PairName = fmt.Sprintf("%s/%s", token1.Symbol, token0.Symbol)
	}

	// 更新下 pair.Type
	if pair.Type == 0 {
		pair.Type = getPairTypeFromChain(pair.Address)
	}

	// 再加上UniV2 or UniV3等尾巴
	if pair.Type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
		pair.PairName = fmt.Sprintf("%s_%s", pair.PairName, "UniV2")
	} else if pair.Type == schema.SWAP_EVENT_UNISWAPV3_LIKE {
		pair.PairName = fmt.Sprintf("%s_%s", pair.PairName, "UniV3")
	}
}

func getPairTypeFromChain(address string) int {
	_type, _ := chain.GetUniPoolType(address)

	return _type
}

func TEST_PAIR() {
	// pair, _ := GetPairInfo("0xEfC97fa9e615D6aE8D4Ed43c14611191f9390ab3", chain.GetMongo())  // 通缩币
	pair, _ := GetPairInfo("0x43bEc83553828f02a4A20BAa536917F709866322", chain.GetMongo()) // normal token
	pair.Type = schema.SWAP_EVENT_UNISWAPV2_LIKE

	_type := chain.GetUniV2PoolTokenHackType(pair)

	utils.Warnf("[ TEST_PAIR ] type: %v", _type)

	// weth, err := chain.GetAccountWEthBalance("0xF97FAB3851F05a3ded46BAf325F58D57405332C3")
	// utils.Warnf("[ test ] weth balance: %v, err: %v", weth, err)
}
