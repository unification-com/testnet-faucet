export GO111MODULE = on

include .env
export $(shell sed 's/=.*//' .env)

build: clean build-front go.sum
	go build -mod=readonly -o build/faucet ./server/faucet.go

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

build-front:
	@cd /tmp && go get github.com/rakyll/statik
	@mkdir -p ./build/client
	@cp -R ./client/public ./build/client/public
	@echo "--> Setting explorer URL to ${MAINCHAIN_EXPLORER_URL}"
	@sed -i 's#__MAINCHAIN_EXPLORER_URL__#${MAINCHAIN_EXPLORER_URL}#g' ./build/client/public/index.html
	@echo "--> Setting ReCaptcha Site Key to ${RECAPTCHA_SITE_KEY}"
	@sed -i 's#__RECAPTCHA_SITE_KEY__#${RECAPTCHA_SITE_KEY}#g' ./build/client/public/index.html
	statik -src=./build/client/public -dest=./client -f -m
	@rm -rf ./build/client

clean:
	rm -rf ./build

fmt:
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s

.PHONY: build build-front clean fmt go.sum
