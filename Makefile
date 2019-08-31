GOCMD=go
GOBUILD=$(GOCMD) build
GOGET=$(GOCMD) get

.PHONY: setup
## Install dependencies
setup:
	$(GOGET) github.com/mitchellh/gox
	$(GOGET) -d -t ./...

.PHONY: cross-build
## Cross build binaries
cross-build:
	rm -rf ./main
	gox -os=linux -arch=amd64 -output=./main -ldflags "-s -w"
