package utils

import (
	"encoding/json"
	"os"
)

type TokenConfig struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	CoingeckoId string `json:"coingeckoId"`
	Address     string `json:"address"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	Category    string `json:"category"`
}

func GetTokenConfig() (map[string]TokenConfig, []string, error) {
	dat, _ := os.ReadFile("./config/tokens/tokens.json")
	var tokenConfigs []TokenConfig
	tokenCgIdList := make([]string, 0)
	if err := json.Unmarshal(dat, &tokenConfigs); err != nil {
		return nil, []string{}, err
	}

	tokenConfigMap := make(map[string]TokenConfig)
	for _, tokenConfig := range tokenConfigs {
		tokenConfigMap[tokenConfig.Symbol] = tokenConfig
		tokenCgIdList = append(tokenCgIdList, tokenConfig.CoingeckoId)
	}

	return tokenConfigMap, tokenCgIdList, nil
}
