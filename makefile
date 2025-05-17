# Generated Makefile for Remake CLI project
.PHONY: all build test clean

all: build

build:
	go build -o bin/remake ./cmd

test:
	go test ./... -v

clean:
	rm -rf bin/ .remake