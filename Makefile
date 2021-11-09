PACKAGES:=$(shell go list ./... | grep -v -e /vendor/)
TAG:=1.0

default: vendor

vendor:
	go mod tidy
	go mod vendor

.PHONY: test
test:
	go test -p=1 -cover -covermode=count $(PACKAGES)

.PHONY: run
run:
# KAFKA_PUBLISHING_PROXY_SSL_CA_LOCATION=$(shell pwd)/testdata
	KAFKA_PUBLISHING_PROXY_APP_CONFIG=$(shell pwd)/.local/app-config.json \
	KAFKA_PUBLISHING_PROXY_SECRETS_PATH=$(shell pwd)/.local \
	KAFKA_PUBLISHING_PROXY_TEMP_DIR=$(shell pwd)/.local \
	go run main.go

.PHONY: docker-run
docker-run:
	docker build -t kpp-test .
	docker run --read-only --rm \
		-p 39000:39000 \
		-v $(shell pwd)/.local:/secrets \
		-e KAFKA_PUBLISHING_PROXY_APP_CONFIG=/secrets/app-config.json \
	 	-e KAFKA_PUBLISHING_PROXY_SECRETS_PATH=/secrets \
		-e KAFKA_PUBLISHING_PROXY_TEMP_DIR=/secrets \
		kpp-test

.PHONY: docker-build
docker-build: test vendor
	docker build -t traviisd/kafka-producer-proxy:$(TAG) .

.PHONY: docker-push
docker-push:
	docker push traviisd/kafka-producer-proxy:$(TAG)