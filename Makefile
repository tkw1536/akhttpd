.PHONY: all deps clean

all: akhttpd

deps:
	go get -v ./...

akhttpd: deps
	go build -o akhttpd ./cmd/akhttpd