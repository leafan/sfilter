package wiser

import (
	"context"
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 用mongo的aggregate命令获取
func GetActiveAccounts(seconds int, mongodb *mongo.Client) ([]string, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.SwapTableName)

	var accounts []string
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "swapTime", Value: bson.D{
					{Key: "$gte", Value: time.Now().Add(-time.Duration(seconds) * time.Second)},
				}},
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$trader"},
			}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second*100)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.Errorf("[ GetActiveAccounts ] aggregate failed: %v", err)
		return accounts, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		utils.Errorf("[ GetActiveAccounts ] cursor failed: %v", err)
		return accounts, err
	}

	for _, result := range results {
		trader, ok := result["_id"].(string)
		if ok && trader != "" {
			accounts = append(accounts, trader)
		}
	}
	// utils.Warnf("[ GetActiveAccounts ] get active accounts len: %v", len(accounts))

	return accounts, nil
}

func GetAccountSwaps(seconds int, pageSize int64, account string, mongodb *mongo.Client) (schema.AccountTrades, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.SwapTableName)

	filter := bson.M{}
	date := time.Now().Add(-time.Duration(seconds) * time.Second)
	filter["swapTime"] = bson.M{
		"$gte": date,
	}
	filter["trader"] = account

	// 升序排列
	options := options.Find().SetSort(bson.D{{Key: "swapTime", Value: 1}})

	trades := make(schema.AccountTrades)

	page := int64(1)
	skip := (page - 1) * pageSize

	for {
		cursor, err := collection.Find(context.Background(), filter, options.SetLimit(pageSize).SetSkip(skip))
		if err != nil {
			return trades, err
		}

		count := 0
		for cursor.Next(context.Background()) {
			var swap schema.Swap
			if err := cursor.Decode(&swap); err != nil {
				cursor.Close(context.Background())
				return trades, err
			}
			count++

			//  针对某一笔swap, 处理出对应数据
			if swap.AmountOfMainToken > 0 {
				att := getAttFromSwap(swap)
				trades[swap.MainToken] = append(trades[swap.MainToken], att)
			}
		}

		if err := cursor.Err(); err != nil {
			cursor.Close(context.Background())
			return trades, err
		}

		if count < int(pageSize) {
			// reached the end of data
			break
		}
		skip += pageSize
	}

	return trades, nil
}

func GetAccountTransfers(seconds int, pageSize int64, account string, mongodb *mongo.Client) (schema.AccountTrades, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TransferTableName)

	filter := bson.M{}
	date := time.Now().Add(-time.Duration(seconds) * time.Second)
	filter["timestamp"] = bson.M{
		"$gte": date,
	}
	filter["$or"] = []bson.M{
		{"from": account},
		{"to": account},
	}

	// 升序排列
	options := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	trades := make(schema.AccountTrades)

	page := int64(1)
	skip := (page - 1) * pageSize

	for {
		cursor, err := collection.Find(context.Background(), filter, options.SetLimit(pageSize).SetSkip(skip))
		if err != nil {
			return trades, err
		}

		count := 0
		for cursor.Next(context.Background()) {
			var transfer schema.Transfer
			if err := cursor.Decode(&transfer); err != nil {
				cursor.Close(context.Background())
				return trades, err
			}
			count++

			//  针对某一笔swap, 处理出对应数据
			if transfer.Amount > 0 {
				att := getAttFromTransfer(transfer, account)
				trades[transfer.Token] = append(trades[transfer.Token], att)
			}
		}

		if err := cursor.Err(); err != nil {
			cursor.Close(context.Background())
			return trades, err
		}

		if count < int(pageSize) {
			// reached the end of data
			break
		}
		skip += pageSize
	}

	return trades, nil
}

func getAttFromSwap(swap schema.Swap) schema.AccountTokenTrade {
	att := schema.AccountTokenTrade{
		BlockNo:  swap.BlockNo,
		TxHash:   swap.TxHash,
		Position: swap.Position,

		Pair:     swap.PairAddr,
		PairType: swap.SwapType,

		TradeTime: swap.SwapTime,

		Type:      schema.TRADE_TYPE_SWAP,
		Direction: swap.Direction,

		Amount:     swap.AmountOfMainToken,
		USDValue:   swap.VolumeInUsd,
		PriceInUSD: swap.PriceInUsd,
	}

	return att
}

func PrintWiser(wiser *schema.Wiser) {
	utils.Infof("**** PrintWiser **** Address: %v", wiser.Address)

	fmt.Println("Weight: ", wiser.Weight)
	fmt.Println("WinRatio: ", wiser.WinRatio)
	fmt.Println("BuyZeroTokenRatio: ", wiser.BuyZeroTokenRatio)
	fmt.Println("ValidTradeCount: ", wiser.ValidTradeCount)
	fmt.Println("TradeCntPerMonth: ", wiser.TradeCntPerMonth)

	fmt.Println("EarnValuePerDeal: ", wiser.EarnValuePerDeal)
	fmt.Println("AverageEarnRatio: ", wiser.AverageEarnRatio)

	fmt.Printf("\n\n")
}

func getAttFromTransfer(transfer schema.Transfer, account string) schema.AccountTokenTrade {
	att := schema.AccountTokenTrade{
		BlockNo:  transfer.BlockNo,
		TxHash:   transfer.TxHash,
		Position: transfer.Position,

		TradeTime: transfer.Timestamp,

		Pair:     "", // transfer没有pair概念
		PairType: schema.SWAP_EVENT_UNKNOWN,
		Type:     schema.TRADE_TYPE_TRANSFER,

		Amount:   transfer.Amount,
		USDValue: transfer.TransferValueInUsd,
	}

	// 方向, 如果to为account, 则为 receive, 否则反之
	if transfer.To == account {
		att.Direction = schema.DIRECTION_BUY_OR_ADD
	} else {
		att.Direction = schema.DIRECTION_SELL_OR_DECREASE
	}

	if att.Amount > 0 {
		att.PriceInUSD = att.USDValue / att.Amount
	} else {
		att.PriceInUSD = 0
	}

	return att
}
