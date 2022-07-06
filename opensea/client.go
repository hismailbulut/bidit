package opensea

import (
	"bidit/bidutil"
	"bidit/opensea/model"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/beefsack/go-rate"
	"github.com/ethereum/go-ethereum/common"
)

const (
	DEFAULT_RETRY_COUNT = 3
	DEFAULT_COOLDOWN    = time.Second
)

var (
	getRate       = rate.New(4, time.Second)
	postRate      = rate.New(2, time.Second)
	openseaGuard  sync.Mutex
	openseaClient struct {
		apiKey  string
		base    string
		testnet bool
	}
)

func Init(testnet bool) {
	openseaGuard.Lock()
	defer openseaGuard.Unlock()
	if testnet {
		openseaClient.base = "https://testnets-api.opensea.io/api/v1"
	} else {
		openseaClient.base = "https://api.opensea.io/api/v1"
	}
	openseaClient.testnet = testnet
}

func SetApiKey(apiKey string) {
	openseaGuard.Lock()
	openseaClient.apiKey = apiKey
	openseaGuard.Unlock()
}

func ApiKey() string {
	openseaGuard.Lock()
	defer openseaGuard.Unlock()
	return openseaClient.apiKey
}

func Base() string {
	openseaGuard.Lock()
	defer openseaGuard.Unlock()
	return openseaClient.base
}

func IsTestnet() bool {
	openseaGuard.Lock()
	defer openseaGuard.Unlock()
	if openseaClient.base == "" {
		panic("opensea is not initialized")
	}
	return openseaClient.testnet
}

func ChainID() int64 {
	if IsTestnet() {
		return 4 // Testnet
	}
	return 1 // Mainnet
}

func requestHeaders() http.Header {
	header := http.Header{}
	header.Add("accept", "application/json")
	header.Add("content-type", "application/json")
	if ApiKey() != "" {
		header.Add("x-api-key", ApiKey())
		// header.Add("x-api-key", "2f6f419a083c46de9d83ce3dbe7db601") // My testnet apikey
	}
	return header
}

func getJson(path []string, params bidutil.Params, v any, retry int, cooldown time.Duration) error {
	getRate.Wait()
	resp, err, code := bidutil.GetRequest(requestHeaders(), Base(), path, params)
	if err != nil {
		fmt.Println("getJson.Err:", err)
		if code == http.StatusTooManyRequests && retry > 0 {
			fmt.Println("Retrying in", cooldown, "Remaining tries:", retry)
			time.Sleep(cooldown)
			return getJson(path, params, v, retry-1, cooldown+time.Second)
		}
		return err
	}
	err = json.Unmarshal(resp, v)
	if err != nil {
		return err
	}
	return nil
}

func postJson(path []string, params bidutil.Params, payload []byte, retry int, cooldown time.Duration) error {
	postRate.Wait()
	_, err, code := bidutil.PostRequest(requestHeaders(), Base(), path, params, payload)
	if err != nil {
		fmt.Println("postJson.Err:", err)
		if code == http.StatusTooManyRequests && retry > 0 {
			fmt.Println("Retrying in", cooldown, "Remaining tries:", retry)
			time.Sleep(cooldown)
			return postJson(path, params, payload, retry-1, cooldown+time.Second)
		}
		return err
	}
	return nil
}

func Asset(contractAddress common.Address, tokenID int) (*model.Asset, error) {
	path := []string{"asset", contractAddress.String(), fmt.Sprint(tokenID)}
	params := bidutil.Params{
		"include_orders": true,
	}
	response := model.Asset{}
	err := getJson(path, params, &response, DEFAULT_RETRY_COUNT, DEFAULT_COOLDOWN)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func Collection(slug string) (*model.Collection, error) {
	path := []string{"collection", slug}
	params := bidutil.Params{}
	response := model.CollectionResponse{
		Success: true, // TODO Check this
	}
	err := getJson(path, params, &response, DEFAULT_RETRY_COUNT, DEFAULT_COOLDOWN)
	if err != nil {
		return nil, err
	}
	return &response.Collection, nil
}
