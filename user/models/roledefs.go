package models

import (
	"sfilter/utils"
	"strconv"
)

const (
	USER_ROLE_BASIC int = iota // 默认, 允许10个地址?

	USER_ROLE_LEVEL_PREMIUM // 允许100个地址
	USER_ROLE_LEVEL_ELITE   // 允许1000个地址

	USER_ROLE_LEVEL_INVESTOR // 允许1000个地址

	// ...

	USER_ROLE_LEVEL_PARTNER = 999 // partner合伙人级别, 允许10000个地址

	USER_ROLE_LEVEL_ROOT = 9999 // root管理员用户
)

// 允许新增地址个数
const (
	ROLE_BASIC_ADDRESS_COUNT    = 10
	ROLE_PREMIUM_ADDRESS_COUNT  = 100
	ROLE_ELITE_ADDRESS_COUNT    = 1000
	ROLE_INVESTOR_ADDRESS_COUNT = 1000
	ROLE_PARTNER_ADDRESS_COUNT  = 10000
	// ...

	ROLE_ROOT_ADDRESS_COUNT = 100000
)

// 允许保存的地址最大数
const (
	ROLE_BASIC_TRACK_SWAP_COUNT    = 1000
	ROLE_PREMIUM_TRACK_SWAP_COUNT  = 1000 * 10
	ROLE_ELITE_TRACK_SWAP_COUNT    = 1000 * 100
	ROLE_INVESTOR_TRACK_SWAP_COUNT = 1000 * 1000
	ROLE_PARTNER_TRACK_SWAP_COUNT  = 1000 * 1000
	// ...

	ROLE_ROOT_TRACK_SWAP_COUNT = 1000 * 1000 * 10
)

func IsValidRole(role int) bool {
	if role == USER_ROLE_BASIC ||
		role == USER_ROLE_LEVEL_PREMIUM ||
		role == USER_ROLE_LEVEL_ELITE ||
		role == USER_ROLE_LEVEL_INVESTOR ||
		role == USER_ROLE_LEVEL_PARTNER {
		return true
	}

	return false
}

func GetRoleTrackAddressCount(roleStr string) int64 {
	role, err := strconv.Atoi(roleStr)
	if err != nil {
		utils.Tracef("[ GetRoleTrackAddressCount ] strconv role(%v) failed: %v", roleStr, err)
		return ROLE_BASIC_ADDRESS_COUNT
	}

	if role == USER_ROLE_LEVEL_PREMIUM {
		return ROLE_PREMIUM_ADDRESS_COUNT
	} else if role == USER_ROLE_LEVEL_ELITE {
		return ROLE_ELITE_ADDRESS_COUNT
	} else if role == USER_ROLE_LEVEL_INVESTOR {
		return ROLE_INVESTOR_ADDRESS_COUNT
	} else if role == USER_ROLE_LEVEL_PARTNER {
		return ROLE_PARTNER_ADDRESS_COUNT
	} else if role == USER_ROLE_LEVEL_ROOT {
		return ROLE_ROOT_ADDRESS_COUNT
	}

	return ROLE_BASIC_ADDRESS_COUNT
}

func GetRoleTrackSwapCount(roleStr string) int64 {
	role, err := strconv.Atoi(roleStr)
	if err != nil {
		return ROLE_BASIC_ADDRESS_COUNT
	}

	if role == USER_ROLE_LEVEL_PREMIUM {
		return ROLE_PREMIUM_TRACK_SWAP_COUNT
	} else if role == USER_ROLE_LEVEL_ELITE {
		return ROLE_ELITE_TRACK_SWAP_COUNT
	} else if role == USER_ROLE_LEVEL_INVESTOR {
		return ROLE_INVESTOR_TRACK_SWAP_COUNT
	} else if role == USER_ROLE_LEVEL_PARTNER {
		return ROLE_PARTNER_TRACK_SWAP_COUNT
	} else if role == USER_ROLE_LEVEL_ROOT {
		return ROLE_ROOT_TRACK_SWAP_COUNT
	}

	return ROLE_BASIC_TRACK_SWAP_COUNT
}
