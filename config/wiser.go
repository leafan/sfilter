package config

type WiserConfig struct {
	WiserSearchInterval int // 每隔多久执行一遍地址更新

	DbBlockReadSize int64 // 每次从db读取的最大数量, 避免卡死db

	AccountActiveSeconds int // 取最近多久有过交易就算活跃地址
	LatestSwapSeconds    int // 判断买卖时, 取最近多久的交易来判断

	ArbitrageBlockInterval int // arbi机器人的买卖间隔区块数
	FrontrunBlockInterval  int // frontrun机器人的买卖间隔区块数
}

var DefaultWiserConfig = &WiserConfig{
	WiserSearchInterval: 60 * 60 * 24,
	DbBlockReadSize:     1000,

	// AccountActiveSeconds:   60 * 60 * 24 * 7,
	AccountActiveSeconds: 60 * 60 * 24, // debug

	LatestSwapSeconds: 60 * 60 * 24 * 2,

	ArbitrageBlockInterval: 0,
	FrontrunBlockInterval:  5,
}
