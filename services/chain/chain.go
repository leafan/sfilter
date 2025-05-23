package chain

import (
	"context"
	"math/big"
	"sfilter/config"
	"sfilter/utils"
	"strings"

	"github.com/cloudfresco/ethblocks"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 本地使用的全局变量, low...
var chainStaticAbi *abi.ABI
var chainStaticBackupAbi *abi.ABI

var staticClient *ethclient.Client
var infuraClient *ethclient.Client

var mongoClient *mongo.Client

func GetMongo() *mongo.Client {
	return getMongo()
}

func GetEthClient() *ethclient.Client {
	return getClient()
}

func getAbi() *abi.ABI {
	if chainStaticAbi == nil {
		abi, err := abi.JSON(strings.NewReader(ChainAbiJson))
		if err != nil {
			utils.Fatalf("getAbi error! err: %v", err)
		}

		chainStaticAbi = &abi
	}

	return chainStaticAbi
}

func getBackupAbi() *abi.ABI {
	if chainStaticBackupAbi == nil {
		abi, err := abi.JSON(strings.NewReader(ChainAbiJsonBackup))
		if err != nil {
			utils.Fatalf("getBackupAbi error! err: %v", err)
		}

		chainStaticBackupAbi = &abi
	}

	return chainStaticBackupAbi
}

func getMongo() *mongo.Client {
	if mongoClient == nil {
		clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
		mongodb, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			utils.Fatalf("getMongo error! err: %v", err)
		}

		mongoClient = mongodb
	}

	return mongoClient
}

func getClient() *ethclient.Client {
	if staticClient == nil {
		cli, err := ethblocks.GetClient(config.WS_ADDR)
		if err != nil {
			utils.Fatalf("getClient error! err: %v", err)
		}

		staticClient = cli
	}

	return staticClient
}

func getInfuraClient() *ethclient.Client {
	if infuraClient == nil {
		cli, err := ethblocks.GetClient(config.INFURA_API_KEY)
		if err != nil {
			utils.Fatalf("infuraClient error! err: %v", err)
		}

		infuraClient = cli
	}

	return infuraClient
}

func _getSingleProp(abi *abi.ABI, address, info string, client *ethclient.Client, height *big.Int) (interface{}, error) {
	contractAddr := common.HexToAddress(address)
	bytes, _ := abi.Pack(info)
	msg := ethereum.CallMsg{
		From: common.Address{},
		To:   &contractAddr,
		Data: bytes,
	}

	ret, err := client.CallContract(context.Background(), msg, height)
	if err != nil {
		// log.Printf("[ getSingleProp ] CallContract error. addr: %v, err: %v\n", address, err)
		return nil, err
	}

	intr, err := abi.Methods[info].Outputs.UnpackValues(ret)
	if err != nil {
		// log.Printf("[ getSingleProp ] UnpackValues error. addr: %v, err: %v\n", address, err)
		return nil, err
	}

	return intr[0], err
}

func GetCurrentBlockNumber() (uint64, error) {
	client := getClient()

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	return header.Number.Uint64(), nil
}

func GetAccountEthBalance(address string) (float64, error) {
	client := getClient()

	balanceBig, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return 0, err
	}

	blanceBFloat := new(big.Float).SetInt(balanceBig)
	blanceBFloat = blanceBFloat.Quo(blanceBFloat, big.NewFloat(1e18))
	balance, _ := blanceBFloat.Float64()

	return balance, nil
}

func GetAccountWEthBalance(address string) (float64, error) {
	balanceBig, err := BalanceOf(address, config.WETH_ADDRESS)
	if err != nil {
		return 0, err
	}

	blanceBFloat := new(big.Float).SetInt(balanceBig)
	blanceBFloat = blanceBFloat.Quo(blanceBFloat, big.NewFloat(1e18))
	balance, _ := blanceBFloat.Float64()

	return balance, nil
}

func GetNoParamProp(address, info string) (interface{}, error) {
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
		return nil, err
	}

	intr, err := abi.Methods[info].Outputs.UnpackValues(ret)
	if err != nil {
		return nil, err
	}

	return intr, err
}

func GetSingleProp(address, info string) (interface{}, error) {
	return getSingleProp(address, info, getClient(), nil)
}

func getSingleProp(address, info string, client *ethclient.Client, height *big.Int) (interface{}, error) {
	abi := getAbi()
	return _getSingleProp(abi, address, info, client, height)
}

func getSingleBackupProp(address, info string, client *ethclient.Client, height *big.Int) (interface{}, error) {
	abi := getBackupAbi()
	return _getSingleProp(abi, address, info, client, height)
}

// 判断是否是合约地址
func IsContract(address string) int {
	const CHECK_CONTRACT = "0x4E013d527f23CD7Cb5b08f6A908de68ce6C57C3e"

	abi := getAbi()
	contractAddr := common.HexToAddress(CHECK_CONTRACT)

	data, err := abi.Pack("isContract", common.HexToAddress(address))
	if err != nil {
		utils.Warnf("[ IsContract ] Pack data error. addr: %v, err: %v", address, err)
		return 0
	}
	msg := ethereum.CallMsg{
		From: common.Address{},
		To:   &contractAddr,
		Data: data,
	}

	ret, err := getClient().CallContract(context.Background(), msg, nil)
	if err != nil {
		utils.Debugf("[ IsContract ] CallContract error. addr: %v, err: %v", address, err)
		return 0
	}

	intr, err := abi.Methods["isContract"].Outputs.UnpackValues(ret)
	if err != nil {
		utils.Debugf("[ IsContract ] UnpackValues error. addr: %v, err: %v", address, err)
		return 0
	}

	retB := intr[0].(bool)
	if retB {
		return 1
	}

	return 0
}

func GetBasicCoinPrice(client *ethclient.Client, height *big.Int, chain string) (float64, error) {
	if client == nil {
		client = getClient()
	}

	if chain == "bsc" {
		return GetBnbPriceByBsc(client)
	}

	return GetEthPrice(client, height)
}

// 获取eth价格,
// 如果配置了height, 需要client支持archive查询功能, 可以用infura
func GetEthPrice(client *ethclient.Client, height *big.Int) (float64, error) {
	if height != nil {
		client = getInfuraClient() // 当指定高度时, 则需要去infura上获取
	}

	const ETH_UNI_POOL = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"
	priceSqrt, err := getSingleProp(ETH_UNI_POOL, "slot0", client, height)

	if err != nil {
		utils.Warnf("[ GetEthPrice ] get price error, will retry via infura. error: %v, height: %v\n", err, height)

		client = getInfuraClient() // 出错了, 重新获取一次
		priceSqrt, err = getSingleProp(ETH_UNI_POOL, "slot0", client, height)
		if err != nil {
			utils.Errorf("[ GetEthPrice ] failed in **infura** again! error: %v, height: %v\n", err, height)

			return 0, err
		}
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

	// utils.Debugf("[ GetEthPrice ] block height: %v, price: %v", height, ret)
	return ret, nil
}

func TEST_CHAIN() {
	block, err := GetCurrentBlockNumber()
	utils.Debugf("block: %v, err: %v", block, err)
}

const ChainAbiJson = `[{
	"inputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "constructor"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "owner",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "spender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "Approval",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1",
		"type": "uint256"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "Burn",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1",
		"type": "uint256"
	}],
	"name": "Mint",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0In",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1In",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0Out",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1Out",
		"type": "uint256"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "Swap",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": false,
		"internalType": "uint112",
		"name": "reserve0",
		"type": "uint112"
	}, {
		"indexed": false,
		"internalType": "uint112",
		"name": "reserve1",
		"type": "uint112"
	}],
	"name": "Sync",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "from",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "Transfer",
	"type": "event"
}, {
	"constant": true,
	"inputs": [],
	"name": "DOMAIN_SEPARATOR",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "MINIMUM_LIQUIDITY",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "PERMIT_TYPEHASH",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"name": "allowance",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "spender",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "approve",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"name": "balanceOf",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "burn",
	"outputs": [{
		"internalType": "uint256",
		"name": "amount0",
		"type": "uint256"
	}, {
		"internalType": "uint256",
		"name": "amount1",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "decimals",
	"outputs": [{
		"internalType": "uint8",
		"name": "",
		"type": "uint8"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "factory",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "getReserves",
	"outputs": [{
		"internalType": "uint112",
		"name": "_reserve0",
		"type": "uint112"
	}, {
		"internalType": "uint112",
		"name": "_reserve1",
		"type": "uint112"
	}, {
		"internalType": "uint32",
		"name": "_blockTimestampLast",
		"type": "uint32"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "_token0",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "_token1",
		"type": "address"
	}],
	"name": "initialize",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "kLast",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "mint",
	"outputs": [{
		"internalType": "uint256",
		"name": "liquidity",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "name",
	"outputs": [{
		"internalType": "string",
		"name": "",
		"type": "string"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"name": "nonces",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "owner",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "spender",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}, {
		"internalType": "uint256",
		"name": "deadline",
		"type": "uint256"
	}, {
		"internalType": "uint8",
		"name": "v",
		"type": "uint8"
	}, {
		"internalType": "bytes32",
		"name": "r",
		"type": "bytes32"
	}, {
		"internalType": "bytes32",
		"name": "s",
		"type": "bytes32"
	}],
	"name": "permit",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "price0CumulativeLast",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "price1CumulativeLast",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "skim",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "uint256",
		"name": "amount0Out",
		"type": "uint256"
	}, {
		"internalType": "uint256",
		"name": "amount1Out",
		"type": "uint256"
	}, {
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"internalType": "bytes",
		"name": "data",
		"type": "bytes"
	}],
	"name": "swap",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "symbol",
	"outputs": [{
		"internalType": "string",
		"name": "",
		"type": "string"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [],
	"name": "sync",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "token0",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "token1",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
},

{
	"constant": true,
	"inputs": [],
	"name": "totalSupply",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, 

{
	"inputs": [],
	"name": "slot0",
	"outputs": [{
		"internalType": "uint160",
		"name": "sqrtPriceX96",
		"type": "uint160"
	}, {
		"internalType": "int24",
		"name": "tick",
		"type": "int24"
	}, {
		"internalType": "uint16",
		"name": "observationIndex",
		"type": "uint16"
	}, {
		"internalType": "uint16",
		"name": "observationCardinality",
		"type": "uint16"
	}, {
		"internalType": "uint16",
		"name": "observationCardinalityNext",
		"type": "uint16"
	}, {
		"internalType": "uint8",
		"name": "feeProtocol",
		"type": "uint8"
	}, {
		"internalType": "bool",
		"name": "unlocked",
		"type": "bool"
	}],
	"stateMutability": "view",
	"type": "function"

},

{
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "transfer",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "from",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "transferFrom",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"inputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "constructor"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "owner",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "spender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "Approval",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1",
		"type": "uint256"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "Burn",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1",
		"type": "uint256"
	}],
	"name": "Mint",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "sender",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0In",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1In",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount0Out",
		"type": "uint256"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "amount1Out",
		"type": "uint256"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "Swap",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": false,
		"internalType": "uint112",
		"name": "reserve0",
		"type": "uint112"
	}, {
		"indexed": false,
		"internalType": "uint112",
		"name": "reserve1",
		"type": "uint112"
	}],
	"name": "Sync",
	"type": "event"
}, {
	"anonymous": false,
	"inputs": [{
		"indexed": true,
		"internalType": "address",
		"name": "from",
		"type": "address"
	}, {
		"indexed": true,
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"indexed": false,
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "Transfer",
	"type": "event"
}, {
	"constant": true,
	"inputs": [],
	"name": "DOMAIN_SEPARATOR",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "MINIMUM_LIQUIDITY",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "PERMIT_TYPEHASH",
	"outputs": [{
		"internalType": "bytes32",
		"name": "",
		"type": "bytes32"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"name": "allowance",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "spender",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "approve",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"name": "balanceOf",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "burn",
	"outputs": [{
		"internalType": "uint256",
		"name": "amount0",
		"type": "uint256"
	}, {
		"internalType": "uint256",
		"name": "amount1",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "decimals",
	"outputs": [{
		"internalType": "uint8",
		"name": "",
		"type": "uint8"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "factory",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "getReserves",
	"outputs": [{
		"internalType": "uint112",
		"name": "_reserve0",
		"type": "uint112"
	}, {
		"internalType": "uint112",
		"name": "_reserve1",
		"type": "uint112"
	}, {
		"internalType": "uint32",
		"name": "_blockTimestampLast",
		"type": "uint32"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "_token0",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "_token1",
		"type": "address"
	}],
	"name": "initialize",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "kLast",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "mint",
	"outputs": [{
		"internalType": "uint256",
		"name": "liquidity",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "name",
	"outputs": [{
		"internalType": "string",
		"name": "",
		"type": "string"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"name": "nonces",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "owner",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "spender",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}, {
		"internalType": "uint256",
		"name": "deadline",
		"type": "uint256"
	}, {
		"internalType": "uint8",
		"name": "v",
		"type": "uint8"
	}, {
		"internalType": "bytes32",
		"name": "r",
		"type": "bytes32"
	}, {
		"internalType": "bytes32",
		"name": "s",
		"type": "bytes32"
	}],
	"name": "permit",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "price0CumulativeLast",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "price1CumulativeLast",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}],
	"name": "skim",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "uint256",
		"name": "amount0Out",
		"type": "uint256"
	}, {
		"internalType": "uint256",
		"name": "amount1Out",
		"type": "uint256"
	}, {
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"internalType": "bytes",
		"name": "data",
		"type": "bytes"
	}],
	"name": "swap",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "symbol",
	"outputs": [{
		"internalType": "string",
		"name": "",
		"type": "string"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [],
	"name": "sync",
	"outputs": [],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "token0",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "token1",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": true,
	"inputs": [],
	"name": "totalSupply",
	"outputs": [{
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	}],
	"payable": false,
	"stateMutability": "view",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "transfer",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
}, {
	"constant": false,
	"inputs": [{
		"internalType": "address",
		"name": "from",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "to",
		"type": "address"
	}, {
		"internalType": "uint256",
		"name": "value",
		"type": "uint256"
	}],
	"name": "transferFrom",
	"outputs": [{
		"internalType": "bool",
		"name": "",
		"type": "bool"
	}],
	"payable": false,
	"stateMutability": "nonpayable",
	"type": "function"
},
{
	"inputs": [{
			"internalType": "address[]",
			"name": "pairs",
			"type": "address[]"
		},
		{
			"internalType": "address[2][]",
			"name": "pairTokens",
			"type": "address[2][]"
		},
		{
			"internalType": "address",
			"name": "inputCurrency",
			"type": "address"
		},
		{
			"internalType": "uint256",
			"name": "amountIn",
			"type": "uint256"
		},
		{
			"internalType": "uint256[2]",
			"name": "checks",
			"type": "uint256[2]"
		}
	],
	"name": "arbitrage",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
},



{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"isContract","outputs":[{"internalType":"bool","name":"totally","type":"bool"}],"stateMutability":"view","type":"function"},


{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount","type":"uint128"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"}],"name":"Burn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"address","name":"recipient","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount0","type":"uint128"},{"indexed":false,"internalType":"uint128","name":"amount1","type":"uint128"}],"name":"Collect","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint128","name":"amount0","type":"uint128"},{"indexed":false,"internalType":"uint128","name":"amount1","type":"uint128"}],"name":"CollectProtocol","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"paid0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"paid1","type":"uint256"}],"name":"Flash","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint16","name":"observationCardinalityNextOld","type":"uint16"},{"indexed":false,"internalType":"uint16","name":"observationCardinalityNextNew","type":"uint16"}],"name":"IncreaseObservationCardinalityNext","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"indexed":false,"internalType":"int24","name":"tick","type":"int24"}],"name":"Initialize","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"int24","name":"tickLower","type":"int24"},{"indexed":true,"internalType":"int24","name":"tickUpper","type":"int24"},{"indexed":false,"internalType":"uint128","name":"amount","type":"uint128"},{"indexed":false,"internalType":"uint256","name":"amount0","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount1","type":"uint256"}],"name":"Mint","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint8","name":"feeProtocol0Old","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol1Old","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol0New","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"feeProtocol1New","type":"uint8"}],"name":"SetFeeProtocol","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"int256","name":"amount0","type":"int256"},{"indexed":false,"internalType":"int256","name":"amount1","type":"int256"},{"indexed":false,"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"indexed":false,"internalType":"uint128","name":"liquidity","type":"uint128"},{"indexed":false,"internalType":"int24","name":"tick","type":"int24"}],"name":"Swap","type":"event"},{"inputs":[{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount","type":"uint128"}],"name":"burn","outputs":[{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount0Requested","type":"uint128"},{"internalType":"uint128","name":"amount1Requested","type":"uint128"}],"name":"collect","outputs":[{"internalType":"uint128","name":"amount0","type":"uint128"},{"internalType":"uint128","name":"amount1","type":"uint128"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint128","name":"amount0Requested","type":"uint128"},{"internalType":"uint128","name":"amount1Requested","type":"uint128"}],"name":"collectProtocol","outputs":[{"internalType":"uint128","name":"amount0","type":"uint128"},{"internalType":"uint128","name":"amount1","type":"uint128"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"factory","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"fee","outputs":[{"internalType":"uint24","name":"","type":"uint24"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"feeGrowthGlobal0X128","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"feeGrowthGlobal1X128","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"flash","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint16","name":"observationCardinalityNext","type":"uint16"}],"name":"increaseObservationCardinalityNext","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"liquidity","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"maxLiquidityPerTick","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"amount","type":"uint128"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"mint","outputs":[{"internalType":"uint256","name":"amount0","type":"uint256"},{"internalType":"uint256","name":"amount1","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"observations","outputs":[{"internalType":"uint32","name":"blockTimestamp","type":"uint32"},{"internalType":"int56","name":"tickCumulative","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityCumulativeX128","type":"uint160"},{"internalType":"bool","name":"initialized","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint32[]","name":"secondsAgos","type":"uint32[]"}],"name":"observe","outputs":[{"internalType":"int56[]","name":"tickCumulatives","type":"int56[]"},{"internalType":"uint160[]","name":"secondsPerLiquidityCumulativeX128s","type":"uint160[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"positions","outputs":[{"internalType":"uint128","name":"liquidity","type":"uint128"},{"internalType":"uint256","name":"feeGrowthInside0LastX128","type":"uint256"},{"internalType":"uint256","name":"feeGrowthInside1LastX128","type":"uint256"},{"internalType":"uint128","name":"tokensOwed0","type":"uint128"},{"internalType":"uint128","name":"tokensOwed1","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"protocolFees","outputs":[{"internalType":"uint128","name":"token0","type":"uint128"},{"internalType":"uint128","name":"token1","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint8","name":"feeProtocol0","type":"uint8"},{"internalType":"uint8","name":"feeProtocol1","type":"uint8"}],"name":"setFeeProtocol","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"slot0","outputs":[{"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"internalType":"int24","name":"tick","type":"int24"},{"internalType":"uint16","name":"observationIndex","type":"uint16"},{"internalType":"uint16","name":"observationCardinality","type":"uint16"},{"internalType":"uint16","name":"observationCardinalityNext","type":"uint16"},{"internalType":"uint8","name":"feeProtocol","type":"uint8"},{"internalType":"bool","name":"unlocked","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"}],"name":"snapshotCumulativesInside","outputs":[{"internalType":"int56","name":"tickCumulativeInside","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityInsideX128","type":"uint160"},{"internalType":"uint32","name":"secondsInside","type":"uint32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"bool","name":"zeroForOne","type":"bool"},{"internalType":"int256","name":"amountSpecified","type":"int256"},{"internalType":"uint160","name":"sqrtPriceLimitX96","type":"uint160"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"swap","outputs":[{"internalType":"int256","name":"amount0","type":"int256"},{"internalType":"int256","name":"amount1","type":"int256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"int16","name":"","type":"int16"}],"name":"tickBitmap","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"tickSpacing","outputs":[{"internalType":"int24","name":"","type":"int24"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"int24","name":"","type":"int24"}],"name":"ticks","outputs":[{"internalType":"uint128","name":"liquidityGross","type":"uint128"},{"internalType":"int128","name":"liquidityNet","type":"int128"},{"internalType":"uint256","name":"feeGrowthOutside0X128","type":"uint256"},{"internalType":"uint256","name":"feeGrowthOutside1X128","type":"uint256"},{"internalType":"int56","name":"tickCumulativeOutside","type":"int56"},{"internalType":"uint160","name":"secondsPerLiquidityOutsideX128","type":"uint160"},{"internalType":"uint32","name":"secondsOutside","type":"uint32"},{"internalType":"bool","name":"initialized","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},


{
	"inputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}, {
		"internalType": "address",
		"name": "",
		"type": "address"
	}, {
		"internalType": "uint24",
		"name": "",
		"type": "uint24"
	}],
	"name": "getPool",
	"outputs": [{
		"internalType": "address",
		"name": "",
		"type": "address"
	}],
	"stateMutability": "view",
	"type": "function"
},

{
	"inputs": [],
	"name": "maxLiquidityPerTick",
	"outputs": [{
		"internalType": "uint128",
		"name": "",
		"type": "uint128"
	}],
	"stateMutability": "view",
	"type": "function"
},


{"inputs":[{"internalType":"address","name":"_factory","type":"address"},{"internalType":"address","name":"_WETH9","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"WETH9","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"factory","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes","name":"path","type":"bytes"},{"internalType":"uint256","name":"amountIn","type":"uint256"}],"name":"quoteExactInput","outputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"tokenIn","type":"address"},{"internalType":"address","name":"tokenOut","type":"address"},{"internalType":"uint24","name":"fee","type":"uint24"},{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint160","name":"sqrtPriceLimitX96","type":"uint160"}],"name":"quoteExactInputSingle","outputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes","name":"path","type":"bytes"},{"internalType":"uint256","name":"amountOut","type":"uint256"}],"name":"quoteExactOutput","outputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"tokenIn","type":"address"},{"internalType":"address","name":"tokenOut","type":"address"},{"internalType":"uint24","name":"fee","type":"uint24"},{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint160","name":"sqrtPriceLimitX96","type":"uint160"}],"name":"quoteExactOutputSingle","outputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"int256","name":"amount0Delta","type":"int256"},{"internalType":"int256","name":"amount1Delta","type":"int256"},{"internalType":"bytes","name":"path","type":"bytes"}],"name":"uniswapV3SwapCallback","outputs":[],"stateMutability":"view","type":"function"},


{
	"inputs": [
		{
			"internalType": "address",
			"name": "pair",
			"type": "address"
		},
		{
			"internalType": "address",
			"name": "from",
			"type": "address"
		},
		{
			"internalType": "address",
			"name": "to",
			"type": "address"
		},
		{
			"internalType": "address",
			"name": "token0",
			"type": "address"
		},
		{
			"internalType": "uint256",
			"name": "amount",
			"type": "uint256"
		},
		{
			"internalType": "uint256",
			"name": "divFactor",
			"type": "uint256"
		}
	],
	"name": "hackTestForUniV2",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
}


]`

const ChainAbiJsonBackup = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"bytes32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"stop","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"guy","type":"address"},{"name":"wad","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"owner_","type":"address"}],"name":"setOwner","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"src","type":"address"},{"name":"dst","type":"address"},{"name":"wad","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"base","type":"uint256"},{"name":"exponent","type":"uint256"}],"name":"pow","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[{"name":"src","type":"address"}],"name":"isOwner","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"guy","type":"address"},{"name":"wad","type":"uint256"}],"name":"mint","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"name_","type":"bytes32"}],"name":"setName","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"src","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"stopped","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"authority_","type":"address"}],"name":"setAuthority","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"bytes32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"guy","type":"address"},{"name":"wad","type":"uint256"}],"name":"burn","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"guy","type":"address"}],"name":"approvex","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"dst","type":"address"},{"name":"wad","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"start","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"authority","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"src","type":"address"},{"name":"guy","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[{"name":"symbol_","type":"bytes32"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"authority","type":"address"}],"name":"LogSetAuthority","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"}],"name":"LogSetOwner","type":"event"},{"anonymous":true,"inputs":[{"indexed":true,"name":"sig","type":"bytes4"},{"indexed":true,"name":"guy","type":"address"},{"indexed":true,"name":"foo","type":"bytes32"},{"indexed":true,"name":"bar","type":"bytes32"},{"indexed":false,"name":"wad","type":"uint256"},{"indexed":false,"name":"fax","type":"bytes"}],"name":"LogNote","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":true,"name":"guy","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"src","type":"address"},{"indexed":true,"name":"dst","type":"address"},{"indexed":false,"name":"wad","type":"uint256"}],"name":"Transfer","type":"event"}

]`
