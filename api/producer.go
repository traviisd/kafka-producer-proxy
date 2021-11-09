package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog"
)

// Result .
type Result struct {
	Message string
	Error   error
}

// ProduceOptions .
type ProduceOptions struct {
	Context context.Context
	Log     *zerolog.Logger
	Cluster string
	Topic   string
	Key     string
	Data    map[string]interface{}
}

type kafkaProducer interface {
	Produce(options ProduceOptions) *Result
}

type producer struct{}

func newProducer() kafkaProducer {
	return &producer{}
}

// Produce publishes the message to Kafka.
func (p producer) Produce(options ProduceOptions) *Result {
	instance, err := getProducer(options.Context, options.Cluster)
	if err != nil {
		return &Result{
			Message: "Could not retrieve Kafka instance.",
			Error:   err,
		}
	}

	ac, _ := kafka.NewAdminClientFromProducer(instance)
	md, err := ac.GetMetadata(&options.Topic, false, 10000)
	if err != nil {
		return &Result{
			Error: err,
		}
	} else if strings.Contains(strings.ToLower(md.Topics[options.Topic].Error.String()), "unknown") {
		return &Result{
			Error: errors.New(md.Topics[options.Topic].Error.String()),
		}
	}

	// used to notify when complete
	done := make(chan *Result)

	defer close(done)

	// event watcher
	go func() {
		for e := range instance.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				done <- &Result{
					Message: fmt.Sprintf("%v", ev.TopicPartition),
					Error:   ev.TopicPartition.Error,
				}
				return
			case *kafka.Error:
				if ev.IsFatal() {
					done <- &Result{
						Error: fmt.Errorf("fatal: %s", ev.Error()),
					}
					return
				}
			default:
				if Config.Debug {
					options.Log.Debug().Msgf("event: %+v", ev)
				}
			}
		}
	}()

	// parse data to byte
	value, err := json.Marshal(options.Data)
	if err != nil {
		return &Result{
			Message: "Could not parse 'data' field",
			Error:   err,
		}
	}

	// send the message
	instance.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &options.Topic,
			Partition: int32(kafka.PartitionAny),
		},
		Key:   []byte(options.Key),
		Value: value,
	}

	return <-done
}
