package main

import (
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"net/http"
)

func main() {
	server := handlers.NewHttpServer()

	err := http.ListenAndServe(":8080", server.Mux)
	if err != nil {
		panic(err)
	}
}
