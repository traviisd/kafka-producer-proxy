package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func Serve() {
	hostname, _ := os.Hostname()
	producer := newProducer()
	router := mux.NewRouter()

	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("appname", "kafka-producer-proxy").
		Str("host", hostname).
		Logger()

	// Install the logger handler with default output on the console
	router.Use(hlog.NewHandler(log))

	// Provided extra handler to set some request's context fields.
	// All logs will come with some prepopulated fields.
	router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	router.Use(hlog.RemoteAddrHandler("ip"))
	router.Use(hlog.UserAgentHandler("user_agent"))
	router.Use(hlog.RefererHandler("referer"))
	router.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

	km, err := newKafkaMiddleware()
	if err != nil {
		log.Fatal().Err(err).Msg("Kafka Init Error")
	}
	router.Use(km.Handler)

	configureRouter(router, producer)

	address := fmt.Sprintf(":%v", Config.ServerPort)

	hs := &http.Server{
		Addr: address,
		// set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	log.Info().Msgf("hostname %s - listening on port %s", hostname, address)

	if Config.EnableTLS {
		// To generate a development cert and key, run the following from your *nix terminal:
		// go run $GOROOT/src/crypto/tls/generate_cert.go --host="localhost"
		panic(hs.ListenAndServeTLS(Config.TLSCert, Config.TLSKey))
	} else {
		panic(hs.ListenAndServe())
	}
}
