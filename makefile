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

test:
	go test ./... -v

coverage: test
	@echo "Generando reporte de cobertura..."
	go test ./... -coverprofile=coverage.out
	@echo "Cobertura total: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"
	@echo "Puedes ver el informe HTML con: go tool cover -html=coverage.out"

lint:
	@echo "Formateando código..."
	go fmt ./...
	@echo "Ejecutando go vet..."
	go vet ./...
	@echo "Ejecutando golangci-lint..."
	golangci-lint run

clean:
	rm -rf bin/ .remake coverage.out
