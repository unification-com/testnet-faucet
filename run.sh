#!/bin/bash
export GOROOT="/home/deploy/go"
export GOPATH="/home/deploy/.go"
export GO111MODULE="on"

/home/deploy/.go/bin/statik -src=./client/public -dest=./client -f -m
make clean
/home/deploy/go/bin/go build -mod=readonly -o build/faucet ./server/faucet.go

build/faucet
