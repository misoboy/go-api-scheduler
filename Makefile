COMMIT_SHA ?= $(shell git rev-parse HEAD)
REPONAME ?= misoboy
IMAGE_NAME ?= go-api-scheduler
DOCKER_TAG ?= latest

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)

LD_FLAGS ?=

.PHONY: build
build:
	go build -o .build/${GOOS}-${GOARCH}/go-api-scheduler ./cmd/api-scheduler

.PHONY: amd64
amd64:
	make GOARCH=amd64 build

.PHONY: arm64
arm64:
	make GOARCH=arm64 build

.PHONY: build-all
build-all: amd64 arm64

.PHONY: build-and-push-api-scheduler
build-and-push-api-scheduler:
	@echo "------------------"
	@echo  "--> Build and push api scheduler docker image"
	@echo "------------------"
	docker buildx build --platform linux/amd64,linux/arm64 --progress plain \
		--no-cache --push -f cmd/api-scheduler/Dockerfile \
		--tag $(REPONAME)/$(IMAGE_NAME):$(DOCKER_TAG) .