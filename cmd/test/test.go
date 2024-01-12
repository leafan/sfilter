package main

import (
	"sfilter/services/chain"
	"sfilter/utils"
)

func test() {
	utils.Infof("****** Debug start ******\n\n")

	chain.TEST_TOKEN()
	// services.TEST_EMAIL_AWS()
	// tutils.TEST_ENCRYPT()

	utils.Infof("****** Debug end  ******\n\n\n")
}

func main() {
	test()
}
