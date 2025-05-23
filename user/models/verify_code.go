package models

import (
	"context"
	"sfilter/config"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VerifyCodeModel struct {
	Collection *mongo.Collection
}

func (m *VerifyCodeModel) GetCodeByUsername(username string, deadline time.Time) (*VerifyCode, error) {
	filter := bson.M{"username": username}
	filter["createdAt"] = bson.M{
		"$gte": deadline,
	}

	var result VerifyCode
	err := m.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	utils.Infof("[ VerifyCodeModel GetCodeByUsername ] find one code for user( %v ) success. code: %v", username, result.Code)
	return &result, nil
}

// 根据ip查找他历史发送的验证码次数
// 目的是防止短期内大量发送浪费资源
func (m *VerifyCodeModel) GetCodesByIp(clientIp string, deadline time.Time) ([]VerifyCode, error) {
	filter := bson.M{"clientIp": clientIp}
	filter["createdAt"] = bson.M{
		"$gte": deadline,
	}
	options := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(10)

	var result []VerifyCode
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := m.Collection.Find(ctx, filter, options)
	if err != nil {
		return result, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &result)
	return result, err
}

func (m *VerifyCodeModel) CreatCode(c *VerifyCode) error {
	c.CreatedAt = time.Now()

	_, err := m.Collection.InsertOne(context.Background(), c)
	if err != nil {
		utils.Errorf("[ VerifyCodeModel CreatCode ] InsertOne error: %v, verifycode: %v\n", err, c)
	}

	utils.Infof("[ VerifyCodeModel CreatCode ] creat code for user( %v ) success.", c.Username)

	return err
}
