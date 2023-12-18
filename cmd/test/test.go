package main

import (
	"sfilter/utils"
)

func test() {
	utils.Infof("****** Debug start ******\n\n")

	// chain.TEST_POOL()
	// services.TEST_EMAIL_AWS()
	utils.TEST_HASH_PASSWORD("deepeye@admin")

	utils.Infof("****** Debug end  ******\n\n\n")
}

func main() {
	test()
}
