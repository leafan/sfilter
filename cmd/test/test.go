package main

import(
	"sfilter/user/auth"
	"sfilter/utils"
)


func test() {
	utils.Infof("****** Debug start ******\n\n")

	// chain.TEST_POOL()
	auth.TEST_EMAIL()

	utils.Infof("****** Debug end  ******\n\n\n")
}

func main() {
	test()
}