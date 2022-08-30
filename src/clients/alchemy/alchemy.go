package alchemy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	alchemyUrl     = "https://eth-mainnet.g.alchemy.com/v2/dVblG1Tfi-psOdZvwsZy_T4mxBHaA-e_"
	zeroBalanceStr = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

type RequestBody struct {
	JsonRpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Params  []string          `json:"params"`
	Id      int               `json:"id"`
}

func GetBalanceOfAnAddress(addr string) {}

func GetTokenBalancesOfAnAddress(addr string) (map[string]string, error) {
	requestBody := RequestBody{
		JsonRpc: "2.0",
		Method:  "alchemy_getTokenBalances",
		Id:      42,
		Params:  []string{addr, "DEFAULT_TOKENS"},
	}

	body, _ := json.Marshal(requestBody)

	resp, err := http.Post(alchemyUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	bt := res["result"].(map[string]interface{})

	tokenBalances := bt["tokenBalances"].([]interface{})

	balances := make(map[string]string)
	for _, balance := range tokenBalances {
		balanceMap, _ := balance.(map[string]interface{})
		balanceStr := fmt.Sprintf("%v", balanceMap["tokenBalance"])
		conractStr := fmt.Sprintf("%v", balanceMap["contractAddress"])
		if balanceStr != zeroBalanceStr {
			fmt.Println(conractStr)
			balances[conractStr] = balanceStr
		}

	}
	return balances, nil
}
