package controllerloops

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/l2planet/l2planet-api/src/clients/coingecko"
	"github.com/l2planet/l2planet-api/src/clients/ethereum"
)

type BridgeConfig struct {
	Address string   `yaml:"address"`
	Tokens  []string `yaml:"tokens"`
}

type ChainConfig struct {
	Bridges []BridgeConfig `yaml:"bridges"`
	Name    string         `yaml:"name"`
}

type TokenConfig struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	CoingeckoId string `json:"coingeckoId"`
	Address     string `json:"address"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	Category    string `json:"category"`
}

func getChainConfig() (ChainConfig, error) {
	dat, _ := os.ReadFile("./config/chains/arbitrum.yaml")
	var chainConfig ChainConfig
	if err := yaml.Unmarshal(dat, &chainConfig); err != nil {
		return ChainConfig{}, err
	}

	return chainConfig, nil
}

func getTokenConfig() (map[string]TokenConfig, []string, error) {
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

// TODO: instead of querying blockchain one by one, use multicall
func CalculateTvl() {
	chainConfig, _ := getChainConfig()
	tokenConfig, cgSymbolList, _ := getTokenConfig()
	tvl := big.NewFloat(0.00)
	ethClient := ethereum.NewClient("https://mainnet.infura.io/v3/0730731b37714d7bafe59025a5cbe455")
	coinGeckoClient := coingecko.NewClient()
	prices, _ := coinGeckoClient.GetPrices(cgSymbolList)

	for _, bridge := range chainConfig.Bridges {
		if len(bridge.Tokens) == 0 {
			for name, _ := range tokenConfig {
				balance, _ := ethClient.BalanceOf(bridge.Address, tokenConfig[name].Address)
				fbalance := new(big.Float)
				fbalance.SetString(balance.String())
				tokenValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(tokenConfig[name].Decimals)))
				coingeckoId := tokenConfig[name].CoingeckoId
				price := (*prices)[coingeckoId]["usd"]

				bigPrice := big.NewFloat(float64(price))
				value := bigPrice.Mul(bigPrice, tokenValue)
				tvl = tvl.Add(tvl, value)
			}
			continue
		}

		for _, tokenName := range bridge.Tokens {
			if tokenName == "ETH" {
				balance, err := ethClient.BalanceAt(bridge.Address)
				if err != nil {
					fmt.Println(err)
					return
				}
				fbalance := new(big.Float)
				fbalance.SetString(balance.String())
				ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
				price := (*prices)["ethereum"]["usd"]
				bigPrice := big.NewFloat(float64(price))
				value := bigPrice.Mul(bigPrice, ethValue)
				tvl = tvl.Add(tvl, value)
			} else {
				balance, err := ethClient.BalanceOf(bridge.Address, tokenConfig[tokenName].Address)
				if err != nil {
					fmt.Println(err)
					return
				}
				fbalance := new(big.Float)
				fbalance.SetString(balance.String())
				tokenValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(tokenConfig[tokenName].Decimals)))
				coingeckoId := tokenConfig[tokenName].CoingeckoId
				price := (*prices)[coingeckoId]["usd"]
				bigPrice := big.NewFloat(float64(price))
				value := bigPrice.Mul(bigPrice, tokenValue)
				tvl = tvl.Add(tvl, value)
			}
		}
	}
	fmt.Println(tvl)
}
