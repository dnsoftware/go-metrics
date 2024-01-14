package app

import (
	"flag"
	"github.com/dnsoftware/go-metrics/internal/server/collector"
	"github.com/dnsoftware/go-metrics/internal/server/handlers"
	"github.com/dnsoftware/go-metrics/internal/storage"
	"net/http"
)

func ServerRun() {

	endpoint := flag.String("a", "localhost:8080", "server endpoint")
	flag.Parse()

	repository := storage.NewMemStorage()

	collect := collector.NewCollector(&repository)

	server := handlers.NewHTTPServer(collect)
	err := http.ListenAndServe(*endpoint, server.Router)
	if err != nil {
		panic(err)
	}

}
