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
func GetActiveAccounts(seconds int, pageSize int64, debugAccount string, mongodb *mongo.Client) (schema.ActiveAccount, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.SwapTableName)

	filter := bson.M{}
	date := time.Now().Add(-time.Duration(seconds) * time.Second)
	filter["swapTime"] = bson.M{
		"$gte": date,
	}

	//  如果定义了 debug account, 则只找对应account数据
	if debugAccount != "" {
		utils.Infof("[ GetActiveAccounts ] debug account now: %v", debugAccount)
		filter["trader"] = debugAccount
	}

	// 升序排列
	options := options.Find().SetSort(bson.D{{Key: "swapTime", Value: 1}})

	accounts := make(schema.ActiveAccount)

	page := int64(1)
	skip := (page - 1) * pageSize
	for {
		cursor, err := collection.Find(context.Background(), filter, options.SetLimit(pageSize).SetSkip(skip))
		if err != nil {
			return accounts, err
		}

		count := 0
		for cursor.Next(context.Background()) {
			var swap schema.Swap
			if err := cursor.Decode(&swap); err != nil {
				cursor.Close(context.Background())
				return accounts, err
			}
			count++

			//  针对某一笔swap, 处理出对应数据
			if swap.Trader != "" {
				if !utils.Contains(accounts[swap.Trader], swap.MainToken) {
					accounts[swap.Trader] = append(accounts[swap.Trader], swap.MainToken)
				}
			}
		}

		if err := cursor.Err(); err != nil {
			cursor.Close(context.Background())
			return accounts, err
		}

		if count < int(pageSize) {
			// reached the end of data
			break
		}
		skip += pageSize
	}

	return accounts, nil
}

func GetAccountSwapsByToken(seconds int, pageSize int64, account, token string, mongodb *mongo.Client) ([]schema.AccountTokenTrade, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.SwapTableName)

	filter := bson.M{}
	date := time.Now().Add(-time.Duration(seconds) * time.Second)
	filter["swapTime"] = bson.M{
		"$gte": date,
	}
	filter["trader"] = account
	filter["mainToken"] = token

	// 升序排列
	options := options.Find().SetSort(bson.D{{Key: "swapTime", Value: 1}})

	var atts []schema.AccountTokenTrade

	page := int64(1)
	skip := (page - 1) * pageSize

	for {
		cursor, err := collection.Find(context.Background(), filter, options.SetLimit(pageSize).SetSkip(skip))
		if err != nil {
			return atts, err
		}

		count := 0
		for cursor.Next(context.Background()) {
			var swap schema.Swap
			if err := cursor.Decode(&swap); err != nil {
				cursor.Close(context.Background())
				return atts, err
			}
			count++

			//  针对某一笔swap, 处理出对应数据
			if swap.AmountOfMainToken > 0 {
				att := getAttFromSwap(swap)
				atts = append(atts, att)
			}
		}

		if err := cursor.Err(); err != nil {
			cursor.Close(context.Background())
			return atts, err
		}

		if count < int(pageSize) {
			// reached the end of data
			break
		}
		skip += pageSize
	}

	return atts, nil
}

func GetAccountTransfersByToken(seconds int, pageSize int64, account, token string, mongodb *mongo.Client) ([]schema.AccountTokenTrade, error) {
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
	filter["token"] = token

	// 升序排列
	options := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	var atts []schema.AccountTokenTrade

	page := int64(1)
	skip := (page - 1) * pageSize

	for {
		cursor, err := collection.Find(context.Background(), filter, options.SetLimit(pageSize).SetSkip(skip))
		if err != nil {
			return atts, err
		}

		count := 0
		for cursor.Next(context.Background()) {
			var transfer schema.Transfer
			if err := cursor.Decode(&transfer); err != nil {
				cursor.Close(context.Background())
				return atts, err
			}
			count++

			//  针对某一笔swap, 处理出对应数据
			if transfer.Amount > 0 {
				att := getAttFromTransfer(transfer, account)
				atts = append(atts, att)
			}
		}

		if err := cursor.Err(); err != nil {
			cursor.Close(context.Background())
			return atts, err
		}

		if count < int(pageSize) {
			// reached the end of data
			break
		}
		skip += pageSize
	}

	return atts, nil
}

func PrintDeals(deals []*schema.BiDeal) {
	for _, deal := range deals {
		PrintDeal(deal)
	}
}

func PrintDeal(deal *schema.BiDeal) {
	utils.Infof("\t**** [ printDeal ] **** Account: %v, TokenName: %v", deal.Account, deal.TokenName)

	fmt.Println("Token: ", deal.Token)
	fmt.Println("BuyTxHash: ", deal.BuyTxHash)
	fmt.Println("BuyBlockNo: ", deal.BuyBlockNo)
	fmt.Println("BuyValue: ", deal.BuyValue)
	fmt.Println("SellTxHashWithToken: ", deal.SellTxHashWithToken)
	fmt.Println("SellBlockNo: ", deal.SellBlockNo)
	fmt.Println("SellValue: ", deal.SellValue)
	fmt.Println("sellType: ", deal.SellType)
	fmt.Println("Earn: ", deal.Earn)
	fmt.Println("EarnChange: ", deal.EarnChange*100, "%")
	fmt.Println("HoldBlocks: ", deal.HoldBlocks)
	fmt.Println("BiDealType: ", deal.BiDealType)

	fmt.Printf("\n\n")
}

func getAttFromSwap(swap schema.Swap) schema.AccountTokenTrade {
	att := schema.AccountTokenTrade{
		BlockNo:  swap.BlockNo,
		TxHash:   swap.TxHash,
		Position: swap.Position,

		Type:      schema.WISER_TYPE_SWAP,
		Direction: swap.Direction,

		Amount:     swap.AmountOfMainToken,
		USDValue:   swap.VolumeInUsd,
		PriceInUSD: swap.PriceInUsd,
	}

	return att
}

func getAttFromTransfer(transfer schema.Transfer, account string) schema.AccountTokenTrade {
	att := schema.AccountTokenTrade{
		BlockNo:  transfer.BlockNo,
		TxHash:   transfer.TxHash,
		Position: transfer.Position,

		Type: schema.WISER_TYPE_TRANSFER,

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
