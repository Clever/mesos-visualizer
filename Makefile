.PHONY: all test build clean
SHELL := /bin/bash
PKGS = $(shell GO15VENDOREXPERIMENT=1 go list ./... | grep -v "vendor/")
GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif

export GO15VENDOREXPERIMENT=1

all: build test run

$(GOPATH)/bin/golint:
	@go get github.com/golang/lint/golint

test: $(PKGS)

$(PKGS): $(GOPATH)/bin/golint
	@echo ""
	@echo "Formatting $@..."
	@gofmt -w=true $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "Linting $@..."
	@$(GOPATH)/bin/golint $@
	@echo ""
	@echo "Testing $@..."
	@go test -v $@

build:
	go build -o mesos-visualizer

run: build
	./mesos-visualizer
