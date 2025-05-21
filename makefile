GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
	GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: all build install test coverage lint clean

all: build

build:
	go build -o bin/remake .

install: build
	@echo "Installing remake to $(GOBIN)"
	mkdir -p $(GOBIN)
	mv bin/remake $(GOBIN)
	rmdir bin
	rm -rf $(HOME)/.remake

test:
	go test ./... -v

coverage: test
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out
	@echo "Coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"
	go tool cover -html=coverage.out

lint:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Running go vet..."
	go vet ./...
	@echo "Running golangci-lint..."
	golangci-lint run

clean:
	rm -rf bin/ .remake coverage.out
