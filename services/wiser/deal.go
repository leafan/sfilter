package wiser

import (
	"context"
	"fmt"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetBiDeals(findOpt *options.FindOptions, filter *primitive.M, mongodb *mongo.Database) ([]schema.BiDeal, int64, error) {
	collection := mongodb.Collection(config.BiDealTableName)

	var result []schema.BiDeal
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, findOpt)
	if err != nil {
		return result, 0, err
	}
	defer cursor.Close(ctx)

	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	totalCount, err := collection.CountDocuments(ctx, filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetBiDeals ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}

func GetAccountAllDeals(account string, mongodb *mongo.Client) ([]schema.BiDeal, error) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiDealTableName)

	var result []schema.BiDeal
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second*10)
	defer cancel()

	filter := bson.M{"account": account}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &result)
	return result, err
}

func SaveDeal(deal *schema.BiDeal, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiDealTableName)

	deal.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), deal)
	if err != nil {
		// 这里失败是正常现象, 重新写入
		filter := bson.D{{Key: "sellTxHashWithToken", Value: deal.SellTxHashWithToken}}
		opts := options.Update().SetUpsert(true)

		update := bson.D{
			{Key: "$set", Value: deal},
		}
		_, err := collection.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			utils.Errorf("[ SaveDeal ] failed. sell_hash: %v, err: %v\n", deal.SellTxHashWithToken, err)
		}

	}
}

// 统计前重置collection
func ResetDealCollection(mongodb *mongo.Client) error {
	collection := mongodb.Database(config.DatabaseName).Collection(config.BiDealTableName)

	err := collection.Drop(context.Background())
	if err != nil {
		utils.Errorf("[ ResetDealCollection ] drop failed: %v", err)
		return err
	}

	utils.DoInitTable(config.DatabaseName, config.BiDealTableName, schema.BiDealIndexModel, mongodb)

	return nil
}

func PrintDeals(deals []*schema.BiDeal) {
	for _, deal := range deals {
		PrintDeal(deal)
	}
}

func PrintDeal(deal *schema.BiDeal) {
	utils.Infof("**** PrintDeal **** Account: %v, TokenName: %v", deal.Account, deal.TokenName)

	fmt.Println("Token: ", deal.Token)
	fmt.Println("BuyTxHash: ", deal.BuyTxHash)
	fmt.Println("BuyPair: ", deal.BuyPair)
	fmt.Println("BuyPairAge: ", deal.BuyPairAge)

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
