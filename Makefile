NAME     := mcp-sqlite
VERSION  := $(shell git describe --tags 2>/dev/null)
REVISION := $(shell git rev-parse --short HEAD 2>/dev/null)
SRCS    := $(shell find . -type f -name '*.go' -o -name 'go.*')
LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\""
DOCKER_TAG := cnosuke/$(NAME)

bin/$(NAME): $(SRCS)
	CGO_ENABLED=1 go build -tags timetzdata $(LDFLAGS) -o bin/$(NAME) main.go

.PHONY: test deps inspect clean build-for-linux-amd64 docker-build docker-push docker-all

# For Docker build, we do NOT cross-compile with CGO, we build native in the container
build-for-linux-amd64:
	CGO_ENABLED=1 go build -tags timetzdata $(LDFLAGS) -o bin/$(NAME)-linux-amd64 main.go

deps:
	go mod download

inspect:
	golangci-lint run

clean:
	rm -rf bin/* dist/*

test:
	go test -v ./...

docker-build:
	docker build -t $(DOCKER_TAG):latest .
	@if [ -n "$(VERSION)" ]; then \
		docker tag $(DOCKER_TAG):latest $(DOCKER_TAG):$(VERSION); \
		echo "Tagged: $(DOCKER_TAG):$(VERSION)"; \
	fi
	@echo "Built: $(DOCKER_TAG):latest"

docker-push:
	docker push $(DOCKER_TAG):latest
	@echo "Pushed: $(DOCKER_TAG):latest"
	@if [ -n "$(VERSION)" ]; then \
		docker push $(DOCKER_TAG):$(VERSION); \
		echo "Pushed: $(DOCKER_TAG):$(VERSION)"; \
	fi

docker-all: docker-build docker-push
	@echo "All docker tasks completed."
