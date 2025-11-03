package main

import (
	"log"
	"net/http"

	"github.com/khatibomar/kaedama/internal/api"
	"github.com/khatibomar/kaedama/internal/config"
	"github.com/khatibomar/kaedama/internal/service/proxy"
)

func main() {
	proxyService := proxy.New()
	handler := api.New(proxyService)

	address := ":" + config.Port

	err := http.ListenAndServe(address, handler)
	if err != nil {
		log.Fatal(err)
	}
}
