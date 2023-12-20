package main

import (
	tutils "sfilter/api/utils"
	"sfilter/utils"
)

func test() {
	utils.Infof("****** Debug start ******\n\n")

	// chain.TEST_POOL()
	// services.TEST_EMAIL_AWS()
	tutils.TEST_ENCRYPT()

	utils.Infof("****** Debug end  ******\n\n\n")
}

func main() {
	test()
}
