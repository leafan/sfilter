package handler

import (
	"context"
	"sfilter/config"
	"sfilter/schema"
	"sfilter/services/chain"
	"sfilter/services/pair"
	"sfilter/services/token"
	"sfilter/utils"
	"time"

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
		utils.Fatalf("[ initMaps ] GetTokenMap failed: %v", err)
	}

	s.Pairs, err = pair.GetPairMap(s.Config.DbBlockReadSize, s.DB.Database(config.DatabaseName))
	if err != nil {
		utils.Fatalf("[ initMaps ] GetPairMap failed: %v", err)
	}

	utils.Infof("[ initMaps ] finished initMaps. tokens len: %v, pairs len: %v", len(s.Tokens), len(s.Pairs))
}

// 前置工作等
func (s *Setting) doWiserPreparation() {
	// 检查是否为通缩币、坑人币等
	s.checkPairValidation()
}

func (s *Setting) checkPairValidation() {
	for _, _pair := range s.Pairs {
		if _pair.MainTokenHackType == schema.PAIR_MAINTOKEN_HACK_TYPE_UNINIT ||
			s.Config.ForceUpdatePairHackStatus {
			// 先判断是否有 pairType
			if _pair.Type == 0 {
				_type, err := chain.GetUniPoolType(_pair.Address)
				if err != nil || _type == 0 {
					// utils.Warnf("[ checkPairValidation ] GetUniPoolType failed. pair: %v, _type: %v, err: %v", _pair.Address, _type, err)

					// 说明pair不对, 标记为unknown
					_pair.MainTokenHackType = schema.PAIR_MAINTOKEN_HACK_TYPE_UNKNOWN
					continue
				}
				_pair.Type = _type
			}

			if _pair.Type == schema.SWAP_EVENT_UNISWAPV2_LIKE {
				_pair.MainTokenHackType = chain.GetUniV2PoolTokenHackType(_pair)
			} else {
				_pair.MainTokenHackType = schema.PAIR_MAINTOKEN_HACK_TYPE_UNKNOWN
			}

			utils.Infof("[ checkPairValidation ] GetAndUpdatePairHackType. pair: %v, pairType: %v, hackType: %v", _pair.Address, _pair.Type, _pair.MainTokenHackType)

			// 保存进db
			_pair.UpdatedAt = time.Now()
			pair.UpsertPair(_pair, s.DB)
		}
	}

	utils.Infof("[ checkPairValidation ] finished checkPairValidation.")
}
