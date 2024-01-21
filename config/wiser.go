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

	// Deal BuyType 设置
	DealBuyTypeMev    int
	DealBuyTypeFresh  int
	DealBuyTypeSubNew int

	// Wiser统计设置
	DealProfitTarget    float64 // 盈利超过该值才算该笔交易盈利
	WinRatioTarget      float64 // 胜率要求
	DealThresholdPerMon float64 // 每个月交易频率底线, 超过该值才算活跃用户

	WiserMinimumEthBalance float64 // eth最少余额才算有效地址

	DealDefiniteWin  float64 // 盈利多少倍, 无论其是否有交易, 都结算
	DealDefiniteLoss float64 // 亏损多少倍, 无论其是否有交易, 都结算

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

	DealBuyTypeMev:    0,
	DealBuyTypeFresh:  60 * 5,
	DealBuyTypeSubNew: 60 * 60 * 24 * 7,

	DealProfitTarget:    1.0, // 盈利超过x%才算有效盈利
	WinRatioTarget:      0.6, // 胜率超过60%才算胜利一笔
	DealThresholdPerMon: 5,   // 每月至少x笔交易才算

	WiserMinimumEthBalance: 0.1,

	DealDefiniteWin:  5,   // 盈利超过x倍, 直接结算
	DealDefiniteLoss: 0.2, // 亏损超过x%, 直接结算
}
