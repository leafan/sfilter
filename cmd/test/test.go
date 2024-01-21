package main

import (
	"sfilter/services/chain"
	"sfilter/utils"
)

func test() {
	utils.Infof("****** Debug start ******\n\n")

	chain.TEST_POOL()

	// chain.TEST_CHAIN()
	// tutils.TEST_ENCRYPT()

	utils.Infof("****** Debug end  ******\n\n\n")
}

func main() {
	test()
}
