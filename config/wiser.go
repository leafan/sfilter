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

	ValidTrendTradeRatio float64 // 趋势交易比例

	WiserMinimumEthBalance float64 // eth最少余额才算有效地址

	DealDefiniteWin  float64 // 盈利多少倍, 无论其是否有交易, 都结算
	DealDefiniteLoss float64 // 亏损多少倍, 无论其是否有交易, 都结算

	// for debug or sth..
	DebugAccount              string // 调试账号
	ForceUpdatePairHackStatus bool   // 是否强制更新pair的是否通缩币状态等

	// hot pair config
	HotPairCheckInterval int     // 每隔多久执行一遍更新
	MinPairLiquidity     float64 // 池子最小金额
	MinPairCreatAge      int     // 池子创建时长
	VolumeIncrement      float64 // 交易量增长幅度
	MinPriceIncreament   float64 // 价格增长最小幅度
	MaxPriceIncreament   float64 // 价格增长最大幅度

	HotPairHookUrl string
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

	DealProfitTarget:    0.5, // 盈利超过x%才算有效盈利
	WinRatioTarget:      0.6, // 胜率超过60%才算胜利一笔
	DealThresholdPerMon: 5,   // 每月至少x笔交易才算

	ValidTrendTradeRatio: 0.6,

	WiserMinimumEthBalance: 1e-16,

	DealDefiniteWin:  5,   // 盈利超过x倍, 直接结算
	DealDefiniteLoss: 0.2, // 亏损超过x%, 直接结算

	ForceUpdatePairHackStatus: false,

	HotPairCheckInterval: 60 * 60 * 24,
	MinPairLiquidity:     100000, // 10w u
	MinPairCreatAge:      60 * 60 * 24 * 30,
	VolumeIncrement:      0.5, // 交易量要增长 x%
	MinPriceIncreament:   0.05,
	MaxPriceIncreament:   1.0,

	HotPairHookUrl: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=08871391-4500-4d47-8a0a-480652b2161a", // 热点币策略  robot
	// HotPairHookUrl: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=ef2ee7d0-e90a-4559-b915-3fb9f73233ff",
}
