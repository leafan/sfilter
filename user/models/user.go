package models

import (
	"context"
	"sfilter/config"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserModel struct {
	Collection *mongo.Collection
}

func (m *UserModel) GetUserByNameOrEmail(user string) (*User, error) {
	filter := bson.M{}

	filter["$or"] = []bson.M{
		{"username": user},
		{"email": user},
	}

	var result User
	err := m.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *UserModel) GetUserByApiKey(apiKey string) (*User, error) {
	filter := bson.M{}

	filter["apiKey"] = apiKey

	var result User
	err := m.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *UserModel) GetUserByReferCode(refercode string) (*User, error) {
	filter := bson.M{"referCode": refercode}

	var result User
	err := m.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *UserModel) CreatUser(u *User) error {
	u.UpdatedAt = time.Now()

	_, err := m.Collection.InsertOne(context.Background(), u)
	if err != nil {
		utils.Errorf("[ UserModel CreatUser ] InsertOne error: %v, userinfo: %v\n", err, u)
	}

	utils.Infof("[ UserModel CreatUser ] creat user( %v ) success.", u.Username)

	return err
}

func (m *UserModel) ResetUserPassword(username, passwd string) error {
	info := struct {
		Passwd    string    `bson:"passwd"`
		UpdatedAt time.Time `bson:"updatedAt"`
	}{
		Passwd:    passwd,
		UpdatedAt: time.Now(),
	}

	filter := bson.D{{Key: "username", Value: username}}
	update := bson.D{
		{Key: "$set", Value: info},
	}

	_, err := m.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Warnf("[ ResetUserPassword ] failed. username: %v, err: %v\n", username, err)
	} else {
		utils.Infof("[ ResetUserPassword ] success reset pass for user: %v", username)
	}

	return nil
}

func (m *UserModel) ResetUserRole(username string, role int) error {
	filter := bson.D{{Key: "username", Value: username}}
	info := struct {
		Role      int       `bson:"role"`
		UpdatedAt time.Time `bson:"updatedAt"`
	}{
		Role:      role,
		UpdatedAt: time.Now(),
	}

	update := bson.D{
		{Key: "$set", Value: info},
	}

	_, err := m.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Warnf("[ ResetUserRole ] failed. username: %v, err: %v\n", username, err)
	} else {
		utils.Infof("[ ResetUserRole ] success reset role for user: %v", username)
	}

	return nil
}

func (m *UserModel) UpdateUserApiKey(username, apiKey string) error {
	filter := bson.D{{Key: "username", Value: username}}
	info := struct {
		ApiKey    string    `bson:"apiKey"`
		UpdatedAt time.Time `bson:"updatedAt"`
	}{
		ApiKey:    apiKey,
		UpdatedAt: time.Now(),
	}

	update := bson.D{
		{Key: "$set", Value: info},
	}

	_, err := m.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Warnf("[ UpdateUserApiKey ] update failed. username: %v, err: %v\n", username, err)
	} else {
		utils.Infof("[ UpdateUserApiKey ] udpate success for user: %v", username)
	}

	return nil
}

func (m *UserModel) GetAllUsers(findOpts *options.FindOptions, filter *primitive.M) ([]User, int64, error) {
	var result []User
	ctx, cancel := context.WithTimeout(context.Background(), config.MONGO_FIND_TIMEOUT*time.Second)
	defer cancel()

	cursor, err := m.Collection.Find(ctx, filter, findOpts)
	if err != nil {
		return result, 0, err
	}
	defer cursor.Close(ctx)

	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	totalCount, err := m.Collection.CountDocuments(ctx, filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetAllUsers ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}
