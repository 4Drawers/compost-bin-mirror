package main

import (
	"compost-bin/router"
)

func main() {
	router := router.Latest()
	router.Logger.Fatal(router.Start(":17890"))
}
