package utils

import "regexp"

func IsValidEthereumAddress(address string) bool {
	ethereumAddressRegex := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

	return ethereumAddressRegex.MatchString(address)
}
