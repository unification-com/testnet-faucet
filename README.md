# Testnet Faucet

Go script for serving the UND Testnet Faucet

## Configure

1. edit `client/public/index.html` and replace Recaptcha key in the `data-sitekey` attribute in the `div.g-recaptcha` container
2. Copy `.env.example` to `.env` and edit (see section [env Variables](#env-Variables) below)

## Build

Statik is required to build the front-end client:

```bash
cd /tmp && go get github.com/rakyll/statik
```

1. run `make build-front` to rebuild the `statik` data
2. run `make clean && make build`

## Running

Run:

```bash
./build/faucet
```

Faucet front end will be accessible via http://[FAUCET_PUBLIC_URL], e.g. http://localhost:8000

### env Variables

`CHAIN_ID`: ID of chain to connect to, e.g. `UND-Mainchain-DevNet`  
`RECAPTCHA_SECRET_KEY`: Secret key from Recaptcha Admin CP  
`FAUCET_AMOUNT_TO_SEND`: amount to send with each request, e.g. 10000000000  
`FAUCET_DENOM`: denomination of coin, e.g. nund  
`NODE_KEY_NAME`: moniker/account name used in keychain to identify faucet sending account  
`NODE_KEY_PASS`: password for sending account's key  
`FAUCET_NODE_RPC_URL`: URL for node RPC which will process `undcli` send command e.g. tcp://localhost:26661  
`FAUCET_PUBLIC_URL`: URL to serve Faucet without protocol prefix, e.g. localhost:8000  
`MAINCHAIN_EXPLORER_URL`: Explorer URL
`FACUET_UNDCLI_HOME`: home dir for faucet to use when calling `undcli` (passed
via the `--home` flag)
`GO_BIN_DIR`: /path/to/go/bin (where `undcli` is installed)
