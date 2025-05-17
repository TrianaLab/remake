# Generated Makefile for Remake CLI project

GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
    GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: all build install test clean

all: build

build:
	go build -o bin/remake .

install: build
	@echo "Installing remake to $(GOBIN)"
	mkdir -p $(GOBIN)
	mv bin/remake $(GOBIN)

test:
	go test ./... -v

clean:
	rm -rf bin/ .remake