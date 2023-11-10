package schema

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func CheckSwapEvent(topics []common.Hash) int {
	if isUniswapSwapV2Event(topics) {
		return SWAP_EVENT_UNISWAPV2_LIKE
	}

	if isUniswapSwapV3Event(topics) {
		return SWAP_EVENT_UNISWAPV3_LIKE
	}

	return SWAP_EVENT_UNKNOWN
}

func isUniswapSwapV2Event(topics []common.Hash) bool {
	if len(topics) == 3 {
		return strings.EqualFold(topics[0].String(), "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822")
	}

	return false
}

func isUniswapSwapV3Event(topics []common.Hash) bool {
	if len(topics) == 3 {
		return strings.EqualFold(topics[0].String(), "0xbeee1e6e7fe307ddcf84b0a16137a4430ad5e2480fc4f4a8e250ab56ccd7630d")
	}

	return false
}
