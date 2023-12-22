package models

import "strconv"

const (
	USER_ROLE_BASIC int = iota // 默认, 允许10个地址?

	USER_ROLE_LEVEL_PREMIUM // 允许100个地址
	USER_ROLE_LEVEL_ELITE   // 允许1000个地址

	USER_ROLE_LEVEL_INVESTOR // 允许10000个地址

	// ...

	USER_ROLE_LEVEL_ROOT = 9999 // root管理员用户
)

const (
	ROLE_BASIC_ADDRESS_COUNT    = 10
	ROLE_PREMIUM_ADDRESS_COUNT  = 100
	ROLE_ELITE_ADDRESS_COUNT    = 1000
	ROLE_INVESTOR_ADDRESS_COUNT = 10000

	// ...

	ROLE_ROOT_ADDRESS_COUNT = 100000
)

func GetRoleTrackCount(roleStr string) int64 {
	role, err := strconv.Atoi(roleStr)
	if err != nil {
		return ROLE_BASIC_ADDRESS_COUNT
	}

	if role == USER_ROLE_LEVEL_PREMIUM {
		return ROLE_PREMIUM_ADDRESS_COUNT
	} else if role == USER_ROLE_LEVEL_ELITE {
		return ROLE_ELITE_ADDRESS_COUNT
	} else if role == USER_ROLE_LEVEL_INVESTOR {
		return ROLE_INVESTOR_ADDRESS_COUNT
	} else if role == USER_ROLE_LEVEL_ROOT {
		return ROLE_ROOT_ADDRESS_COUNT
	}

	return ROLE_BASIC_ADDRESS_COUNT
}
