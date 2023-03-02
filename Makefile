version=$(shell git describe --always --dirty=-dirty)

.PHONY: all 

all:	lint build

lint:
	gofmt -w *.go cmd/unifi_exporter/*.go

build:
	go build ./cmd/unifi_exporter

docker:
	# docker buildx create --name mybuilder
	docker buildx build --platform linux/arm64,linux/amd64 --push -t visago/unifi_exporter:${version} .
