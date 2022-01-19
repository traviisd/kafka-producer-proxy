package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Router is the main interface
type Router interface {
	Ping(w http.ResponseWriter, r *http.Request)
	Health(w http.ResponseWriter, r *http.Request)
	PublishEvent(w http.ResponseWriter, r *http.Request)
	GetAvailableClusters(w http.ResponseWriter, r *http.Request)
}

type router struct {
	kp kafkaProducer
}

// configureRouter returns a new instance of Router
func configureRouter(mr *mux.Router, kp kafkaProducer) {
	r := &router{kp}

	mr.HandleFunc("/ping", r.Ping).Methods(http.MethodGet)
	mr.HandleFunc("/health", r.Health).Methods(http.MethodGet)
	mr.HandleFunc("/events", r.PublishEvent).Methods(http.MethodPost, http.MethodDelete)
	mr.HandleFunc("/clusters", r.GetAvailableClusters).Methods(http.MethodGet)
}

// Health checks the health of the API. Should try
// to run commands to ensure proper permissions.
func (rh router) Health(w http.ResponseWriter, r *http.Request) {
	errs := &bytes.Buffer{}

	for _, cluster := range Config.KafkaBrokerGroups {
		var p *kafka.Producer
		var ac *kafka.AdminClient
		var err error

		p, err = getProducer(r.Context(), cluster)
		if err != nil {
			errs.WriteString(fmt.Sprintf("%s\n", err.Error()))
		}

		if p != nil {
			ac, err = kafka.NewAdminClientFromProducer(p)
			if err != nil {
				errs.WriteString(fmt.Sprintf("%s\n", err.Error()))
			}
		}

		if ac != nil {
			// Get a single or all topics, timeout of 15 seconds
			_, err = ac.GetMetadata(Config.KafkaHealthTopic, (Config.KafkaHealthTopic == nil), 15000)
			if err != nil {
				errs.WriteString(fmt.Sprintf("%s\n", err.Error()))
			}
		}
	}

	if errs.Len() > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"errors\": \"%s\"}", errs.String())))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// Ping is just a means to indicate the API is up and running and ready for traffic.
func (rh router) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// GetAvailableClusters returns the list of available brokers defined in the app-config.json
func (rh router) GetAvailableClusters(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(map[string][]string{
		"clusters": Config.KafkaBrokerGroups,
	})

	if err != nil {
		writeErrorResponse(w, hlog.FromRequest(r), "", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

type EventRequest struct {
	Cluster string                 `json:"cluster"`
	Topic   string                 `json:"topic"`
	Key     interface{}            `json:"key"`
	Data    map[string]interface{} `json:"data"`
}

type eventResponse struct {
	Message string `json:"message,omitempty"`
}

func (rh router) PublishEvent(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)

	// Optional authentication via API token.
	if Config.EnableAPIAuth {
		apiToken := r.Header.Get("X-API-TOKEN")
		if !strings.EqualFold(Secrets.APIToken, apiToken) {
			writeErrorResponse(w, log, fmt.Sprintf("API Token Request Failed: RemoteAddress %s", r.RemoteAddr), nil)
			return
		}
	}

	var er EventRequest

	body, _ := ioutil.ReadAll(r.Body)

	if err := json.Unmarshal(body, &er); err != nil {
		writeErrorResponse(w, log, "error deserializing request body", err)
		return
	}

	log.Debug().Msg(fmt.Sprintf("%s: %s", r.Method, er))

	result := rh.kp.Produce(ProduceOptions{
		Context: r.Context(),
		Log:     log,
		Cluster: er.Cluster,
		Topic:   er.Topic,
		Key:     er.Key,
		Data:    er.Data,
	})

	if result.Error != nil {
		writeErrorResponse(w, log, "", result.Error)
		return
	}

	b, _ := json.Marshal(eventResponse{result.Message})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b))
}

type errorResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func writeErrorResponse(w http.ResponseWriter, log *zerolog.Logger, message string, err error) {
	er := errorResponse{}

	if len(message) > 0 {
		er.Message = message
	}

	if err != nil {
		er.Error = err.Error()
	}

	log.Error().Msgf("%+v", er)

	b, _ := json.Marshal(er)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(b)
}
