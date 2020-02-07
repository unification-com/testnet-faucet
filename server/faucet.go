package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/joho/godotenv"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/tendermint/tmlibs/bech32"
	"github.com/tomasen/realip"

	_ "github.com/unification-com/testnet-faucet/client/statik"
)

type FaucetConfig struct {
	ChainID            string
	ReCaptchaSecretKey string
	FaucetAmountToSend string
	FaucetDenom        string
	NodeKeyName        string
	NodePass           string
	NodeRpcUrl         string
	FaucetPublicUrl    string
	FaucetCliHomeDir   string
	GoBinDir           string
}

type FaucetRequest struct {
	To        string
	Recaptcha string
}

type CliOut struct {
	TxHash string
}

var config FaucetConfig

type FaucetResponse struct {
	Success bool
	Tx      string
	Msg     string
	Amount  string
}

func (fr FaucetResponse) String() string {
	str := strings.TrimSpace(fmt.Sprintf("{\"success\":%v,\"msg\":\"%v\",\"tx\":\"%v\",\"amount\":\"%v\"}", fr.Success, fr.Msg, fr.Tx, fr.Amount))
	return str
}

func procDotEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		fmt.Println(key, "=", value)
		return value
	} else {
		log.Fatal("Error processing .env variable: ", key)
		return ""
	}
}

func main() {
	Init()
	cmd := GetFaucetCmd()

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func GetFaucetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "faucet",
		Short: "Unification Testnet Faucet server application",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("running...")
			fmt.Println(config)
			StartServer()
			return nil
		},
	}
}

func Init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config.ChainID = procDotEnv("CHAIN_ID")
	config.ReCaptchaSecretKey = procDotEnv("RECAPTCHA_SECRET_KEY")
	config.FaucetAmountToSend = procDotEnv("FAUCET_AMOUNT_TO_SEND")
	config.FaucetDenom = procDotEnv("FAUCET_DENOM")
	config.NodeKeyName = procDotEnv("NODE_KEY_NAME")
	config.NodePass = procDotEnv("NODE_KEY_PASS")
	config.NodeRpcUrl = procDotEnv("FAUCET_NODE_RPC_URL")
	config.FaucetPublicUrl = procDotEnv("FAUCET_PUBLIC_URL")
	config.FaucetCliHomeDir = procDotEnv("FACUET_UNDCLI_HOME")
	config.GoBinDir = procDotEnv("GO_BIN_DIR")
}

func StartServer() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	recaptcha.Init(config.ReCaptchaSecretKey)

	http.HandleFunc("/get_nund", faucetSendHandler)
	http.Handle("/", http.FileServer(statikFS))

	fmt.Println("starting faucet server on", config.FaucetPublicUrl)
	if err := http.ListenAndServe(config.FaucetPublicUrl, nil); err != nil {
		log.Fatal("failed to start faucet server", err)
	}
}

func executeCmd(command string, nodePass ...string) (string, error) {
	cmd, wc, out := goExecute(command)

	_, _ = wc.Write([]byte("y\n"))
	time.Sleep(time.Second)

	for _, write := range nodePass {
		_, _ = wc.Write([]byte(write + "\n"))
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(out)
	if err != nil {
		panic(err)
	}

	var cliOut CliOut
	decoder := json.NewDecoder(buf)
	err = decoder.Decode(&cliOut)
	if err != nil {
		panic(err)
	}

	err = cmd.Wait()
	return cliOut.TxHash, err
}

func goExecute(command string) (cmd *exec.Cmd, pipeIn io.WriteCloser, pipeOut io.ReadCloser) {
	cmd = getCmd(command)
	pipeIn, _ = cmd.StdinPipe()
	pipeOut, _ = cmd.StdoutPipe()
	go cmd.Start()
	time.Sleep(time.Second)
	return cmd, pipeIn, pipeOut
}

func getCmd(command string) *exec.Cmd {
	// split command into command and args
	split := strings.Split(command, " ")

	var cmd *exec.Cmd
	if len(split) == 1 {
		cmd = exec.Command(split[0])
	} else {
		cmd = exec.Command(split[0], split[1:]...)
	}

	return cmd
}

func faucetSendHandler(w http.ResponseWriter, req *http.Request) {

	var faucetRequest FaucetRequest
	buf := new(bytes.Buffer)
	ln, err := buf.ReadFrom(req.Body)

	if ln == 0 {
		serveError(w, http.StatusText(400), 400)
		return
	}

	if err != nil {
		serveError(w, err.Error(), 500)
		return
	}

	decoder := json.NewDecoder(buf)
	err = decoder.Decode(&faucetRequest)
	if err != nil {

		fmt.Println(err)
		serveError(w, err.Error(), 500)
		return
	}

	fmt.Println(faucetRequest)

	// make sure address is bech32
	readableAddress, decodedAddress, err := bech32.DecodeAndConvert(faucetRequest.To)
	if err != nil {
		serveError(w, err.Error(), 500)
		return
	}
	// re-encode the address in bech32
	encodedAddress, err := bech32.ConvertAndEncode(readableAddress, decodedAddress)
	if err != nil {
		serveError(w, err.Error(), 500)
		return
	}

	// make sure captcha is valid
	clientIP := realip.FromRequest(req)
	captchaResponse := faucetRequest.Recaptcha
	captchaPassed, err := recaptcha.Confirm(clientIP, captchaResponse)
	if err != nil {
		serveError(w, err.Error(), 403)
		return
	}

	if captchaPassed {

		homeDir := ""
		if len(config.FaucetCliHomeDir) > 0 {
			homeDir = " --home " + config.FaucetCliHomeDir
		}

		undCliCmd := fmt.Sprintf(
			"%vundcli tx send %v %v %v%v --chain-id %v --node %v --gas auto --gas-adjustment 1.5 --gas-prices 0.025nund --memo UND_TestNet_Faucet --output json%v",
			config.GoBinDir, config.NodeKeyName, encodedAddress, config.FaucetAmountToSend, config.FaucetDenom, config.ChainID, config.NodeRpcUrl, homeDir)

		fmt.Println(undCliCmd)

		txHash, err := executeCmd(undCliCmd, config.NodePass)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Println(txHash)

		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")

		// quick hack to return UND amount from nund

		fromAmt, _ := strconv.ParseFloat(config.FaucetAmountToSend, 64)
		fromAmtBf := new(big.Float).SetFloat64(fromAmt)
		res := fromAmtBf.Mul(fromAmtBf, big.NewFloat(1e-9))
		amt := res.Text('f', 0)

		resp := FaucetResponse{
			Success: true,
			Tx:      txHash,
			Amount:  amt,
			Msg:     "",
		}

		fmt.Println("sending", resp)

		js, _ := json.Marshal(resp)

		_, _ = w.Write(js)

		return
	}

	serveError(w, "Captcha failed", 403)
}

func serveError(w http.ResponseWriter, errMsg string, code int) {

	msg := FaucetResponse{
		Success: false,
		Tx:      "",
		Amount:  "0",
		Msg:     errMsg,
	}

	fmt.Println("error", msg)

	js, _ := json.Marshal(msg)
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(js)
}
