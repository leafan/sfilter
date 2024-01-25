package main

import (
	"sfilter/services/pair"
	"sfilter/utils"
)

func test() {
	utils.Infof("****** Debug start ******\n\n")

	pair.TEST_PAIR()

	// chain.TEST_CHAIN()
	// tutils.TEST_ENCRYPT()

	utils.Infof("****** Debug end  ******\n\n\n")
}

func main() {
	test()
}
