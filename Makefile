.PHONY: all 

all:	lint build

lint:
	gofmt -w *.go cmd/unifi_exporter/*.go

build:
	go build ./cmd/unifi_exporter

docker:
	docker build -t visago/unifi_exporter .
