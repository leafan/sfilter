package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserModel struct {
	DB         *mongo.Client
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
