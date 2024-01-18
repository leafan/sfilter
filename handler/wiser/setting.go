package handler

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/pair"
	"sfilter/services/token"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Setting struct {
	DB     *mongo.Client
	Config *config.WiserConfig

	Tokens schema.TokenMap
	Pairs  schema.PairMap
}

func NewSetting(account, db string, debug bool) *Setting {
	clientOptions := options.Client().ApplyURI(config.MONGO_ADDR)
	mongodb, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		utils.Fatalf("connect mongo error: %v", err)
	}

	wiserConfig := config.DefaultWiserConfig
	if account != "" {
		wiserConfig.DebugAccount = account
	}

	if debug {
		wiserConfig.DebugMode = true
	}

	if db != "" {
		config.DatabaseName = db
	}
	// other config change..

	set := &Setting{
		DB:     mongodb,
		Config: wiserConfig,
	}

	set.InitSettings()

	return set
}

func (s *Setting) InitSettings() {
	schema.InitTables(s.DB)

	s.initMaps()
}

func (s *Setting) initMaps() {
	var err error
	s.Tokens, err = token.GetTokenMap(s.Config.DbBlockReadSize, s.DB.Database(config.DatabaseName))
	if err != nil {
		utils.Fatalf("[ InitMap ] GetTokenMap failed: %v", err)
	}

	s.Pairs, err = pair.GetPairMap(s.Config.DbBlockReadSize, s.DB.Database(config.DatabaseName))
	if err != nil {
		utils.Fatalf("[ InitMap ] GetPairMap failed: %v", err)
	}

	utils.Infof("[ InitMap ] tokens len: %v, pairs len: %v", len(s.Tokens), len(s.Pairs))
}
