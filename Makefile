NAME = "ansip"
COMMIT = $(shell git rev-parse --short HEAD)
PARAMS = -v -trimpath -ldflags "-X 'main.version=$(COMMIT)' -s -w -buildid="
MAIN = ./cmd/ansip

.PHONY: build

build:
	go build $(PARAMS) $(MAIN)