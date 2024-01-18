package config

type WiserConfig struct {
	DebugMode           bool
	WiserSearchInterval int // 每隔多久执行一遍地址更新

	DbBlockReadSize int64 // 每次从db读取的最大数量, 避免卡死db
	MongoTimeout    int   // mongo 延迟时间

	AccountActiveSeconds int // 取最近多久有过交易就算活跃地址
	LatestSwapSeconds    int // 判断买卖时, 取最近多久的交易来判断

	ArbitrageBlockInterval int // arbi机器人的买卖间隔区块数
	FrontrunBlockInterval  int // frontrun机器人的买卖间隔区块数
	GambleBlockInterval    int // Game投机交易买卖间隔区块数
	RushBlockInterval      int // 高频交易间隔区块数

	// Wiser统计设置
	ProfitTarget        float64 // 盈利超过该值才算该笔交易盈利
	WinRatioTarget      float64 // 胜率要求
	DealThresholdPerMon float64 // 每个月交易频率底线, 超过该值才算活跃用户

	BotTradeRatio   float64 // frontrun占比超过多少就认为是bot
	HumanTradeRatio float64 // trend交易占比超过多少就认为是human

	// for debug..
	DebugAccount string // 调试账号
}

var DefaultWiserConfig = &WiserConfig{
	WiserSearchInterval: 60 * 60 * 24,
	DbBlockReadSize:     5000,
	MongoTimeout:        60 * 5, // 5min

	AccountActiveSeconds: 60 * 60 * 24 * 7, // 7天内有过交易的用户

	LatestSwapSeconds: 60 * 60 * 24 * 30, // 查询账户的历史买卖

	ArbitrageBlockInterval: 0,
	FrontrunBlockInterval:  5,   // 1min
	GambleBlockInterval:    24,  // 5min
	RushBlockInterval:      128, // 30min

	ProfitTarget:        0.4,  // 盈利超过40%才算有效盈利
	WinRatioTarget:      0.6,  // 胜率超过60%
	DealThresholdPerMon: 10.0, // 每月至少10笔交易才算

	BotTradeRatio:   0.5,
	HumanTradeRatio: 0.5,
}
