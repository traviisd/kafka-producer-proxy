package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

// Secrets are application secrets
var (
	Secrets        *appSecrets
	KafkaCertFiles *kafkaCertFiles
)

type KafkaConfig struct {
	Idempotence      bool   `json:"enable.idempotence"`
	BootstrapServers string `json:"bootstrap.servers"`
	SecurityProtocol string `json:"security.protocol"`
	SaslMechanisms   string `json:"sasl.mechanisms,omitempty"`
	Username         string `json:"sasl.username,omitempty"`
	Password         string `json:"sasl.password,omitempty"`
}

type kafkaCertFiles struct {
	CAChain string
	CerRaw  string
	KeyRaw  string
}

type certConfig struct {
	CAChain        []string `json:"ca_chain"`
	Certificate    string   `json:"certificate"`
	IssuingCA      string   `json:"issuing_ca"`
	PrivateKey     string   `json:"private_key"`
	PrivateKeyType string   `json:"private_key_type"`
}

// Add any secret field definitions here.
type appSecrets struct {
	OAuthClientSecret string `json:"oAuthClientSecret"`
	APIToken          string `json:"apiToken"`
	// Kind of using secrets as a keystore, but this is the simplest way to map clusters with secrets.
	KafkaSecrets map[string]KafkaConfig `json:"kafkaSecrets"`
}

// SetAppSecrets sets the application secrets
func SetAppSecrets(data []byte) {
	if err := json.Unmarshal(data, &Secrets); err != nil {
		log.Err(err).Msg("oops...")
	}
}

// SetCertData sets the application secrets
func SetCertData(data []byte) {
	dir := fmt.Sprintf("%s%s", os.Getenv("KAFKA_PUBLISHING_PROXY_TEMP_DIR"), string(os.PathSeparator))
	cfg := certConfig{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Err(err).Msg("oops...")
	}

	KafkaCertFiles = &kafkaCertFiles{
		CAChain: fmt.Sprintf("%sinternal-ca-chain.pem", dir),
		CerRaw:  cfg.Certificate,
		KeyRaw:  cfg.PrivateKey,
	}

	// create cert chain file
	cabuf := &bytes.Buffer{}
	for _, value := range cfg.CAChain {
		if _, err := cabuf.WriteString(fmt.Sprintf("%s\n", value)); err != nil {
			log.Err(err).Msg("oops...")
		}
	}

	if err := ioutil.WriteFile(KafkaCertFiles.CAChain, cabuf.Bytes(), 0644); err != nil {
		log.Err(err).Msg("oops...")
	}
}

func kafkaClusterLookup(cluster string) (kc KafkaConfig, err error) {
	if value, ok := Secrets.KafkaSecrets[cluster]; ok {
		return value, err
	}

	keys := make([]string, 0, len(Secrets.KafkaSecrets))
	for key := range Secrets.KafkaSecrets {
		keys = append(keys, key)
	}

	return kc, fmt.Errorf("%s not found. Available clusters:\n%v", cluster, strings.Join(keys, ","))
}
