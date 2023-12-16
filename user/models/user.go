package models

import (
	"context"
	"sfilter/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	u.UpdateAt = time.Now()

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
