package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/insolar/insolar/api/requester"
	"github.com/insolar/insolar/log"
	"github.com/spf13/pflag"
)

var (
	apiURL         string
	method         string
	params         string
	memberRef      string
	paramsFile     string
	memberKeysPath string
)

type response struct {
	Jsonrpc string
	id      int64
	Error   interface{}
	Result  interface{}
}

const defaultURL = "http://localhost:19101/api"

func parseInputParams() {
	pflag.StringVarP(&memberKeysPath, "memberkeys", "k", "", "path to file with Member keys")
	pflag.StringVarP(&apiURL, "url", "u", defaultURL, "api url")
	pflag.StringVarP(&paramsFile, "paramsFile", "f", "", "json file params")

	pflag.StringVarP(&method, "method", "m", "", "Insolar method name")
	pflag.StringVarP(&params, "params", "p", "", "params JSON")
	pflag.StringVarP(&memberRef, "address", "i", "", "Insolar member ref")

	pflag.Parse()
}

func check(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func main() {
	parseInputParams()

	err := log.SetLevel("error")
	check("can't set 'error' level on logger: ", err)
	request := &requester.Request{
		JSONRPC: "2.0",
		ID:      0,
		Method:  "api.call",
	}

	if paramsFile != "" {
		request, err = readRequestParams(paramsFile)
		check("[ simpleRequester ]", err)
	} else {
		if memberRef == "" {
			response, err := requester.Info(apiURL)
			check("[ simpleRequester ]", err)
			request.Params.Reference = response.RootMember
		} else {
			request.Params.Reference = memberRef
		}
		if method != "" {
			request.Params.CallSite = method
		}
		if params != "" {
			request.Params.CallParams = params
		}
	}
	if request.Params.CallSite == "" {
		fmt.Println("Method cannot be null", err)
		os.Exit(1)
	}

	seed, err := requester.GetSeed(apiURL)
	check("[ simpleRequester ]", err)
	request.Params.Seed = seed

	rawConf, err := ioutil.ReadFile(memberKeysPath)
	check("[ simpleRequester ]", err)

	stringParams, _ := json.Marshal(request.Params.CallParams)
	fmt.Println("Params: " + string(stringParams))
	fmt.Println("Method: " + request.Method)
	fmt.Println("Reference: " + request.Params.Reference)
	fmt.Println("Seed: " + request.Params.Seed)

	keys := &memberKeys{}
	err = json.Unmarshal(rawConf, keys)
	check("[ simpleRequester ]", err)
	request.Params.PublicKey = keys.Public
	dataToSign, err := json.Marshal(request)
	check("[ simpleRequester ]", err)

	// ks := platformpolicy.NewKeyProcessor()
	// publicKey, err := ks.ImportPublicKeyPEM([]byte(keys.Public))
	// check("[ simpleRequester ]", err)
	signature, err := requester.Sign(keys.Public, dataToSign)
	check("[ simpleRequester ]", err)

	fmt.Println("signature: " + signature)
	body, err := requester.GetResponseBodyContract(apiURL+"/call", *request, signature)
	check("[ simpleRequester ]", err)
	if len(body) == 0 {
		fmt.Println("[ simpleRequester ] Response body is Empty")
	}

	response, err := getResponse(body)
	check("[ simpleRequester ]", err)

	fmt.Println("Execute result: ", response)
}
