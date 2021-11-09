package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traviisd/kafka-producer-proxy/api"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	api.SetAppConfig(os.Getenv("KAFKA_PUBLISHING_PROXY_APP_CONFIG"))

	done := make(chan bool)
	defer func() {
		close(done)
	}()

	go configureSecrets(done)
	// need to warmup so secrets get added.
	time.Sleep(time.Second * 5)

	api.Serve()
}

func configureSecrets(done chan bool) {
	secretsPath := os.Getenv("KAFKA_PUBLISHING_PROXY_SECRETS_PATH")
	sf := fmt.Sprintf("%s/secrets.json", secretsPath)
	kcf := fmt.Sprintf("%s/internal-ca.json", secretsPath)

	files := []api.DynamicFile{
		{
			File:       sf,
			UpdateFunc: api.SetAppSecrets,
		},
	}

	if api.Config.UseKafkaCertAuth {
		files = append(files, api.DynamicFile{
			File:       kcf,
			UpdateFunc: api.SetCertData,
		})
	}

	fpw := api.NewFilePathWatcher(secretsPath, files)

	// set initial secrets
	if err := fpw.UpdateDynamicFile(sf); err != nil {
		time.Sleep(time.Second * 3)
		log.Err(err).Msg("error setting initial secrets.json data, restarting...")
	}

	if api.Config.UseKafkaCertAuth {
		if err := fpw.UpdateDynamicFile(kcf); err != nil {
			time.Sleep(time.Second * 3)
			log.Err(err).Msg("error setting initial internal-ca.json data, restarting...")
		}
	}

	fpw.Watch(done)
}
