package main

import (
	"flag"
	handler "sfilter/handler/wiser"
)

func main() {
	db := flag.String("db", "", "the db want to use")
	account := flag.String("account", "", "the account want to test, empty for all accounts")

	debug := flag.Bool("debug", false, "debug mode")

	deal := flag.Bool("deal", false, "whether enable deal inspect")
	wiser := flag.Bool("wiser", false, "whether enable wiser inspect")

	hx := flag.String("hx", "", "whether enable hot big or etc")

	flag.Parse()

	hndl := handler.NewHandler(*account, *db, *debug, *deal, *wiser, *hx)

	hndl.Run()
}
