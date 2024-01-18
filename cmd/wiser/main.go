package main

import (
	"flag"
	handler "sfilter/handler/wiser"
)

func main() {
	db := flag.String("db", "", "the db want to use")
	debug := flag.Bool("debug", false, "debug mode")
	account := flag.String("account", "", "the account want to test, empty for all accounts")

	wiser := flag.Bool("wiser", true, "whether enable wiser searcher")

	flag.Parse()

	hndl := handler.NewHandler(*account, *db, *debug, *wiser)

	hndl.Run()
}
