module github.com/traviisd/kafka-producer-proxy

go 1.15

replace github.com/traviisd/kafka-producer-proxy/api => ./api

require (
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/howeyc/fsnotify v0.9.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/rs/zerolog v1.26.0
	github.com/stretchr/testify v1.7.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
)
