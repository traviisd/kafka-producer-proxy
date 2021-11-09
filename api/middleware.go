package api

import (
	"context"
	"errors"
	"fmt"

	"net/http"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type contextKey string

var producerctxkey = contextKey("producerctx")

type producerCTX struct {
	Cluster  string `json:"cluster"`
	Instance *kafka.Producer
}

var producerCTXs []producerCTX

// kafkaMiddleware holds functions for setting a  instance to
// Context within http.Handler middleware
type kafkaMiddleware struct{}

// newKafkaMiddleware returns an instance of  producers
func newKafkaMiddleware() (*kafkaMiddleware, error) {
	producerCTXs = []producerCTX{}

	// Create Producer instances
	for _, kc := range Config.KafkaBrokerGroups {
		cfg, err := kafkaClusterLookup(kc)

		kcm := kafka.ConfigMap{
			"enable.idempotence": cfg.Idempotence,
			"bootstrap.servers":  cfg.BootstrapServers,
			"security.protocol":  cfg.SecurityProtocol,
		}

		if Config.Debug {
			kcm["debug"] = "all"
		}

		if Config.UseKafkaCertAuth {
			kcm["ssl.ca.location"] = KafkaCertFiles.CAChain
			kcm["ssl.certificate.pem"] = KafkaCertFiles.CerRaw
			kcm["ssl.key.pem"] = KafkaCertFiles.KeyRaw
		} else {
			kcm["ssl.ca.location"] = os.Getenv("KAFKA_PUBLISHING_PROXY_SSL_CA_LOCATION")
			kcm["sasl.mechanisms"] = cfg.SaslMechanisms
			kcm["sasl.username"] = cfg.Username
			kcm["sasl.password"] = cfg.Password
		}

		kp, err := kafka.NewProducer(&kcm)

		if err != nil {
			return nil, err
		}

		producerCTXs = append(producerCTXs, producerCTX{
			Cluster:  kc,
			Instance: kp,
		})
	}

	return &kafkaMiddleware{}, nil
}

// Handler adds the instance to the request context
func (*kafkaMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), producerctxkey, producerCTXs))
		h.ServeHTTP(w, r)
	})
}

// getProducer retrieves the producer instance from context and Panics if not found
func getProducer(ctx context.Context, cluster string) (*kafka.Producer, error) {
	instance, ok := ctx.Value(producerctxkey).([]producerCTX)
	if !ok {
		return nil, errors.New("kafka producers were not found in context")
	}

	for _, p := range instance {
		if strings.EqualFold(cluster, p.Cluster) {
			return p.Instance, nil
		}
	}

	return nil, fmt.Errorf("kafka producer with the name '%s' was not found", cluster)
}
