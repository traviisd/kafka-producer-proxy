PACKAGES:=$(shell go list ./... | grep -v -e /vendor/)
TAG:=v1.0-beta

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

# make docker-build -e TAG=v1.1
.PHONY: docker-build
docker-build: test vendor
	docker build -t traviisd/kafka-producer-proxy:$(TAG) .

# make docker-push -e TAG=v1.1
.PHONY: docker-push
docker-push:
	docker push traviisd/kafka-producer-proxy:$(TAG)

# make git-tag -e TAG=v1.1
.PHONY: git-tag
git-tag:
	git tag $(TAG)
	git push origin --tags

.PHONY: helm-package
helm-package:
	helm lint .helm
	helm package .helm
	helm repo index --url https://traviisd.github.io/kafka-producer-proxy/ .