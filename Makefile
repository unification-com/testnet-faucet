export GO111MODULE = on

include .env
export $(shell sed 's/=.*//' .env)

build: go.sum
	go build -mod=readonly -o build/faucet ./server/faucet.go

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

build-front:
	@echo "--> Setting explorer URL to ${MAINCHAIN_EXPLORER_URL}"
	@sed -i 's#__MAINCHAIN_EXPLORER_URL__#${MAINCHAIN_EXPLORER_URL}#g' ./client/public/index.html
	@echo "--> Setting ReCaptcha Site Key to ${RECAPTCHA_SITE_KEY}"
	@sed -i 's#__RECAPTCHA_SITE_KEY__#${RECAPTCHA_SITE_KEY}#g' ./client/public/index.html
	statik -src=./client/public -dest=./client -f -m

clean:
	rm -rf ./build

fmt:
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s

.PHONY: build build-front clean fmt go.sum
