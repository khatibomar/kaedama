package main

import (
	"log"
	"net/http"
	"time"

	"github.com/khatibomar/kaedama/internal/api"
	"github.com/khatibomar/kaedama/internal/config"
	"github.com/khatibomar/kaedama/internal/service/proxy"
)

func main() {
	proxyService := proxy.New()
	handler := api.New(proxyService)

	address := ":" + config.Port

	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
