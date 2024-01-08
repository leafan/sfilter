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

type TrackAddressModel struct {
	Collection *mongo.Collection
}

func (m *TrackAddressModel) GetTrackAddressesByUsername(username string, optionsClient *options.FindOptions, filterClient *primitive.M) ([]UserTrackedAddress, int64, error) {
	filter := filterClient
	if filter == nil {
		filter = &bson.M{}
	}
	(*filter)["username"] = username

	findOpts := optionsClient
	if findOpts == nil {
		findOpts = options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetLimit(10)
	}

	var result []UserTrackedAddress
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
		utils.Warnf("[ GetTrackAddressesByUsername ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}

// 判断是否已经存在
func (m *TrackAddressModel) GetEntryByUserAndAddress(username, address string) (*UserTrackedAddress, error) {
	filter := bson.M{"username": username, "address": address}

	var result UserTrackedAddress
	err := m.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *TrackAddressModel) GetEntryByUserAndMemo(username, memo string) (*UserTrackedAddress, error) {
	filter := bson.M{"username": username, "memo": memo}

	var result UserTrackedAddress
	err := m.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *TrackAddressModel) CreatOne(t *UserTrackedAddress) error {
	t.UpdatedAt = time.Now()
	t.CreatedAt = time.Now()

	_, err := m.Collection.InsertOne(context.Background(), t)
	if err != nil {
		utils.Errorf("[ TrackAddressModel CreatOne ] InsertOne error: %v, data: %v\n", err, t)
	}

	return err
}

func (m *TrackAddressModel) GetUserTrackAddressCount(username string) (int64, error) {
	filter := bson.M{"username": username}

	countOpts := &options.CountOptions{
		Limit: &config.COUNT_UPPER_SIZE,
	}
	countOpts.SetMaxTime(config.MONGO_FIND_TIMEOUT * time.Second)

	totalCount, err := m.Collection.CountDocuments(context.Background(), filter, countOpts)
	if err != nil {
		utils.Warnf("[ GetUserTrackAddressCount ] Count error: %v\n", err)
		return 0, err
	}

	return totalCount, nil
}

func (m *TrackAddressModel) UpdateTrackedAddress(username, address string, addrInfo *AddressInfo) error {
	filter := bson.M{"username": username, "address": address}

	info := struct {
		AddressInfo `bson:",inline"`
		UpdatedAt   time.Time `bson:"updatedAt"`
	}{
		AddressInfo: *addrInfo,
		UpdatedAt:   time.Now(),
	}
	update := bson.D{
		{Key: "$set", Value: info},
	}

	_, err := m.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		utils.Warnf("[ UpdateTrackedAddress ] failed. username: %v, address: %v, err: %v\n", username, address, err)
	}

	return err
}

func (m *TrackAddressModel) DeleteByUsernameAndAddr(username, address string) error {
	filter := bson.M{"username": username, "address": address}

	_, err := m.Collection.DeleteOne(context.Background(), filter)
	if err != nil {
		utils.Errorf("[ TrackAddressModel DeleteByUsernameAndAddr ] failed, error: %v, username: %v, address: %v\n", err, username, address)
	}

	return err
}

func (m *TrackAddressModel) GetTrackAddressMap(page, pageSize int64) (UserTrackAddressMap, int, error) {
	umap := make(UserTrackAddressMap)
	var err error
	var totalCount int

	skip := (page - 1) * pageSize
	for {
		cursor, err := m.Collection.Find(context.Background(), bson.M{}, options.Find().SetLimit(pageSize).SetSkip(skip))
		if err != nil {
			return umap, 0, err
		}

		count := 0
		for cursor.Next(context.Background()) {
			var addr UserTrackedAddress
			if err := cursor.Decode(&addr); err != nil {
				cursor.Close(context.Background())
				return umap, 0, err
			}
			count++
			totalCount++

			umap[addr.Address] = append(umap[addr.Address], addr)
		}

		if err := cursor.Err(); err != nil {
			cursor.Close(context.Background())
			return umap, 0, err
		}

		if count < int(pageSize) {
			// reached the end of data
			break
		}
		skip += pageSize
	}

	// utils.Tracef("[ GetTrackAddressMap ]  unique addresses count: %v", len(umap))
	return umap, totalCount, err
}

/////// Admin ////////

func (m *TrackAddressModel) GetAllTrackAddrs(optionsClient *options.FindOptions, filter *primitive.M) ([]UserTrackedAddress, int64, error) {
	findOpts := optionsClient
	if findOpts == nil {
		findOpts = options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetLimit(10)
	}

	var result []UserTrackedAddress
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
		utils.Warnf("[ GetAllTrackAddrs ] Count error: %v\n", err)
		return result, 0, err
	}

	err = cursor.All(ctx, &result)
	return result, totalCount, err
}
