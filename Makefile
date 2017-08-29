include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: all test build clean
SHELL := /bin/bash
PKG := github.com/Clever/mesos-visualizer
PKGS := $(shell go list ./... | grep -v /vendor/)
EXECUTABLE := $(shell basename $(PKG))
$(eval $(call golang-version-check,1.8))

$(GOPATH)/bin/glide:
	@go get github.com/Masterminds/glide

all: test build run

build:
	@CGO_ENABLED=0 go build -a -installsuffix cgo -o $(EXECUTABLE) $(PKG)

run: build
	docker build -t mesos-visualizer .
	@docker run -p 8080:80 \
		-v `pwd`/static:/bin/static/ \
		--env-file=<(echo -e $(_ARKLOC_ENV_FILE)) mesos-visualizer

test: $(PKGS)
$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))

install_deps: $(GOPATH)/bin/glide
	@$(GOPATH)/bin/glide install
