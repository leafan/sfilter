package main

import (
	"flag"
	handler "sfilter/handler/wiser"
)

func main() {
	db := flag.String("db", "", "the db want to use")
	account := flag.String("account", "", "the account want to test, empty for all accounts")
	token := flag.String("token", "", "the token want to test, empty for all tokens")
	flag.Parse()

	wiser := handler.NewWiser(*account, *token, *db)

	wiser.Run()
}
