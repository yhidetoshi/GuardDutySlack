export GO111MODULE=on


## Install dependencies
.PHONY: deps
deps:
	go get -v -d


## Setup development
.PHONY: deps
devel-deps: deps
	GO111MODULE=off
	go get -u golang.org/x/lint/golint
	go get -u github.com/motemen/gobump/cmd/gobump
	go get -u github.com/Songmu/make2help/cmd/make2help


## Setup build
.PHONY: pre-build
build-deps:
	go get -u github.com/mitchellh/gox


## Build binaries
.PHONY: build
build: build-deps
	rm -rf ./main
	gox -os=linux -arch=amd64 -output=./main -ldflags "-s -w"
	gobump show


## Lint
.PHONY: lint
lint: devel-deps
	go vet ./...
	golint -set_exit_status ./...


## Show help
.PHONY: help
help:
	@make2help $(MAKEFILE_LIST)