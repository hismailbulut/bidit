package main

import (
	"bidit/bidutil"
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"
)

type EtherscanCommonResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  T      `json:"result"`
}

type EthPrice struct {
	Ethbtc          decimal.Decimal `json:"ethbtc"`
	EthbtcTimestamp decimal.Decimal `json:"ethbtc_timestamp"`
	Ethusd          decimal.Decimal `json:"ethusd"`
	EthusdTimestamp decimal.Decimal `json:"ethusd_timestamp"`
}

type GasPrice struct {
	LastBlock       decimal.Decimal `json:"LastBlock"`
	SafeGasPrice    decimal.Decimal `json:"SafeGasPrice"`
	ProposeGasPrice decimal.Decimal `json:"ProposeGasPrice"`
	FastGasPrice    decimal.Decimal `json:"FastGasPrice"`
	SuggestBaseFee  decimal.Decimal `json:"suggestBaseFee"`
	GasUsedRatio    string          `json:"gasUsedRatio"`
}

func FetchEthPrice() (*EthPrice, error) {
	resp, err, _ := bidutil.GetRequest(http.Header{}, "https://api.etherscan.io/api", []string{}, bidutil.Params{
		"module": "stats",
		"action": "ethprice",
	})
	if err != nil {
		return nil, err
	}
	response := EtherscanCommonResponse[EthPrice]{}
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	if response.Status != "1" {
		return nil, PrintError(response.Message)
	}
	return &response.Result, nil
}

func FetchGasPrice() (*GasPrice, error) {
	resp, err, _ := bidutil.GetRequest(http.Header{}, "https://api.etherscan.io/api", []string{}, bidutil.Params{
		"module": "gastracker",
		"action": "gasoracle",
	})
	if err != nil {
		return nil, err
	}
	response := EtherscanCommonResponse[GasPrice]{}
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	if response.Status != "1" {
		return nil, PrintError(response.Message)
	}
	return &response.Result, nil
}
