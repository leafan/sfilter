package schema

import (
	"sfilter/config"
	"sfilter/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

func InitTables(mongodb *mongo.Client) {
	utils.DoInitTable(config.DatabaseName, config.SwapTableName, SwapIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.BlockProceededTableName, BlockProceededIndexModel, mongodb)

	utils.DoInitTable(config.DatabaseName, config.TokenTableName, TokenIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.PairTableName, PairIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.LiquidityEventTableName, LiquidityEventIndexModel, mongodb)

	utils.DoInitTable(config.DatabaseName, config.Kline1MinTableName, Kline1MinIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.Kline1HourTableName, Kline1HourIndexModel, mongodb)

	utils.DoInitTable(config.DatabaseName, config.TransferTableName, TransferIndexModel, mongodb)

	utils.DoInitTable(config.DatabaseName, config.ConfigTableName, ConfigIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.GlobalTrendTableName, GlobalTrendIndexModel, mongodb)

	utils.DoInitTable(config.DatabaseName, config.TrackSwapTableName, TrackSwapIndexModel, mongodb)

	// facet
	utils.DoInitTable(config.DatabaseName, config.FacetTableName, FacetIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.InscriptionTableName, InscriptionIndexModel, mongodb)

	// wiser
	utils.DoInitTable(config.DatabaseName, config.WiserTableName, WiserIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.BiDealTableName, BiDealIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.BiTradeTableName, BiTradeIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.HotPairRankTableName, HotPairRankIndexModel, mongodb)

	// router etc
	utils.DoInitTable(config.DatabaseName, config.RouterTableName, RouterIndexModel, mongodb)
	utils.DoInitTable(config.DatabaseName, config.SpecialAddressTableName, SpecialAddressIndexModel, mongodb)
}
