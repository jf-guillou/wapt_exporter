package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	namespace = "wapt"
)

var (
	addr         = flag.String("web.listen-address", ":9976", "The address to listen on for HTTP requests.")
	endpoint     = flag.String("web.endpoint", "/metrics", "Path under which to expose metrics.")
	waptApi      = flag.String("wapt.api", "http://127.0.0.1:8080", "WAPT API endpoint")
	waptUser     = flag.String("wapt.user", "user", "WAPT API username")
	waptPassword = flag.String("wapt.password", "user", "WAPT API password")
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Probably need a flag to handle log levels
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	flag.Parse()

	err := prometheus.Register(NewWaptCollector())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register collector")
	}

	http.Handle(*endpoint, promhttp.Handler())
	log.Fatal().Err(http.ListenAndServe(*addr, nil)).Msg("ListenAndServe")
}
