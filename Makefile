include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: all test build clean
SHELL := /bin/bash
PKG := github.com/Clever/mesos-visualizer
PKGS := $(shell go list ./... | grep -v /vendor/)
EXECUTABLE := $(shell basename $(PKG))
$(eval $(call golang-version-check,1.13))


all: test build run

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o $(EXECUTABLE) $(PKG)

run: build
	docker build -t mesos-visualizer .
	@docker run -p 8080:80 \
		-v `pwd`/static:/bin/static/ \
		-v $(AWS_SHARED_CREDENTIALS_FILE):$(AWS_SHARED_CREDENTIALS_FILE) \
		--env-file=<(echo -e $(_ARKLOC_ENV_FILE)) mesos-visualizer

test: $(PKGS)
$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

install_deps: golang-dep-vendor-deps
	$(call golang-dep-vendor)
