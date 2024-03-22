package main

import (
	"net/http"

	"github.com/enolgor/muxc/examples/basic/controllers"
	"github.com/enolgor/muxc/examples/basic/muxc"
)

func main() {
	mux := http.NewServeMux()
	muxc.ConfigureMux(mux, controllers.NewController())
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
