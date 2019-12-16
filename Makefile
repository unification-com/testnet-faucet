export GO111MODULE = on

build: go.sum
	go build -mod=readonly -o build/faucet ./server/faucet.go

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

build-front:
	statik -src=./client/public -dest=./client -f -m

clean:
	rm -rf ./build

fmt:
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s

.PHONY: build build-front clean fmt go.sum
