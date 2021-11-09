package api

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config is the configuration instance.
var Config appConfig

type appConfig struct {
	Debug             bool     `json:"debug"`
	ServerPort        int      `json:"serverPort"`
	EnableAPIAuth     bool     `json:"enableApiAuth"`
	EnableTLS         bool     `json:"enableTLS"`
	TLSCert           string   `json:"tlsCert"`
	TLSKey            string   `json:"tlsKey"`
	UseKafkaCertAuth  bool     `json:"useKafkaCertAuth"`
	KafkaBrokerGroups []string `json:"kafkaBrokerGroups"`
}

// SetAppConfig deserializes a config.json (any name) file into the config struct to allow access to
// configuration values.
func SetAppConfig(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}

	b, _ := ioutil.ReadFile(file)

	if err := json.Unmarshal(b, &Config); err != nil {
		return err
	}

	return nil
}
