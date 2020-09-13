.PHONY: all deps clean

all: akhttpd

deps:
	go get -v ./...

generate:
	go generate ./...

akhttpd: deps
	go build -o akhttpd ./cmd/akhttpd