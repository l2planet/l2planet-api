package controllerloops

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/l2planet/l2planet-api/src/clients/coingecko"
	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/clients/ethereum"
	"github.com/l2planet/l2planet-api/src/models"
	"github.com/l2planet/l2planet-api/src/multicall"
	"github.com/l2planet/l2planet-api/src/token"
)

const (
	localDir = "./config/"
)

type BridgeConfig struct {
	Address string   `yaml:"address"`
	Tokens  []string `yaml:"tokens"`
}

type ChainConfig struct {
	Bridges     []BridgeConfig `yaml:"bridges"`
	Name        string         `yaml:"name"`
	Tokens      []string       `yaml:"tokens"`
	Description string         `yaml:"description"`
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

func getTokenConfig() (map[string]TokenConfig, []string, error) {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = localDir
	}
	dat, _ := os.ReadFile(configDir + "tokens/tokens.json")
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

func CalculateTvlMulticall() error {
	//ts := time.Now()
	solutionConfigs, _ := db.GetClient().GetSolutionConfig()
	tokenConfig, cgSymbolList, _ := getTokenConfig()
	ethClient := ethereum.NewClient()
	coinGeckoClient := coingecko.NewClient()
	tokenAbi, _ := abi.JSON(strings.NewReader(token.TokenABI))
	multiCalls := make([]multicall.Multicall2Call, 0)
	prices, _ := coinGeckoClient.GetPrices(cgSymbolList)
	for _, solution := range solutionConfigs {
		for _, bridge := range solution.Bridges {
			tvl := big.NewFloat(0.00)
			var bridgeModel models.Bridge
			db.GetClient().First(&bridgeModel, "contract_adress = ?", bridge.ContractAdress)
			//if No tokens specified, go over all of them, else iterate over the specified list
			if len(bridge.SupportedTokens) == 0 {
				for name := range tokenConfig {
					bridgeAddr := common.HexToAddress(bridge.ContractAdress)
					packedData, err := tokenAbi.Pack("balanceOf", bridgeAddr)
					if err != nil {
						continue
					}
					if tokenConfig[name].Address == "" {
						continue
					}
					tokenAddr := common.HexToAddress(tokenConfig[name].Address)
					multiCalls = append(multiCalls, multicall.Multicall2Call{
						Target:   tokenAddr,
						CallData: packedData,
					})

				}
				ethClient.MulticallBalance(multiCalls)
			} else {
				for _, tokenName := range bridge.SupportedTokens {

					balance, err := getBalance(ethClient, bridge.ContractAdress, tokenConfig[tokenName].Address, tokenConfig[tokenName].Decimals)
					if err != nil {
						fmt.Printf("balance of the %s token cannot be found: %v \n", tokenName, err)
						continue
					}

					coingeckoId := tokenConfig[tokenName].CoingeckoId
					price := (*prices)[coingeckoId]["usd"]
					bigPrice := big.NewFloat(float64(price))

					value := bigPrice.Mul(bigPrice, balance)
					tvl = tvl.Add(tvl, value)
				}
			}
			persistedTvl, _ := tvl.Float64()
			fmt.Println(persistedTvl)
		}
	}
	return nil
}

// TODO: instead of querying blockchain one by one, use multicall
func CalculateTvl() error {
	ts := time.Now()
	solutionConfigs, _ := db.GetClient().GetSolutionConfig()
	tokenConfig, cgSymbolList, _ := getTokenConfig()
	ethClient := ethereum.NewClient()
	coinGeckoClient := coingecko.NewClient()

	prices, _ := coinGeckoClient.GetPrices(cgSymbolList)
	tx := db.GetClient().DB.Begin()
	for _, solution := range solutionConfigs {
		for _, bridge := range solution.Bridges {
			tvl := big.NewFloat(0.00)
			var bridgeModel models.Bridge
			db.GetClient().First(&bridgeModel, "contract_adress = ?", bridge.ContractAdress)

			//if No tokens specified, go over all of them, else iterate over the specified list
			if len(bridge.SupportedTokens) == 0 {
				for name := range tokenConfig {

					balance, err := getBalance(ethClient, bridge.ContractAdress, tokenConfig[name].Address, tokenConfig[name].Decimals)
					if err != nil {
						//fmt.Printf("balance of the %s token cannot be found: %v \n", name, err)
						continue
					}

					//Get Price of the asset
					coingeckoId := tokenConfig[name].CoingeckoId
					price := (*prices)[coingeckoId]["usd"]
					bigPrice := big.NewFloat(float64(price))

					//calculate total value
					value := bigPrice.Mul(bigPrice, balance)
					tvl = tvl.Add(tvl, value)
					/*persistedBalance, _ := balance.Float64()
					db.GetClient().Create(&models.Balance{
						Symbol:    name,
						Value:     persistedBalance,
						Timestamp: ts,
						BridgeID:  bridgeModel.ID,
					})*/
				}
			} else {
				for _, tokenName := range bridge.SupportedTokens {

					balance, err := getBalance(ethClient, bridge.ContractAdress, tokenConfig[tokenName].Address, tokenConfig[tokenName].Decimals)
					if err != nil {
						fmt.Printf("balance of the %s token cannot be found: %v \n", tokenName, err)
						continue
					}

					coingeckoId := tokenConfig[tokenName].CoingeckoId
					price := (*prices)[coingeckoId]["usd"]
					bigPrice := big.NewFloat(float64(price))

					value := bigPrice.Mul(bigPrice, balance)
					tvl = tvl.Add(tvl, value)
					/*
						persistedBalance, _ := balance.Float64()
						db.GetClient().Create(&models.Balance{
							Symbol:    tokenName,
							Value:     persistedBalance,
							Timestamp: ts,
							BridgeID:  bridgeModel.ID,
						})
					*/
				}
			}
			persistedTvl, _ := tvl.Float64()
			if err := tx.Create(&models.Tvl{
				Value:     persistedTvl,
				Timestamp: ts,
				BridgeID:  bridgeModel.ID,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func getBalance(ethClient *ethereum.Client, bridgeAddress, tokenAddress string, decimals int) (*big.Float, error) {
	if tokenAddress == "" {
		balance, err := ethClient.BalanceAt(bridgeAddress)
		if err != nil {
			return nil, err
		}

		fbalance := new(big.Float)
		fbalance.SetString(balance.String())
		ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(decimals)))
		return ethValue, nil
	}

	balance, err := ethClient.BalanceOf(bridgeAddress, tokenAddress)
	if err != nil {
		return nil, err
	}

	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	tokenValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(decimals)))

	return tokenValue, nil

}
