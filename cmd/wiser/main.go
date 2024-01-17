package main

import (
	"flag"
	handler "sfilter/handler/wiser"
)

func main() {
	db := flag.String("db", "", "the db want to use")
	account := flag.String("account", "", "the account want to test, empty for all accounts")
	flag.Parse()

	wiser := handler.NewWiser(*account, *db)

	wiser.Run()
}
