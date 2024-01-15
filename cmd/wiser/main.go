package main

import handler "sfilter/handler/wiser"

func main() {
	wiser := handler.NewWiser()

	wiser.Run()
}
