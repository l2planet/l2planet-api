package controllerloops

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/l2planet/l2planet-api/src/clients/coingecko"
	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/clients/ethereum"
	"github.com/l2planet/l2planet-api/src/models"
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

/*
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
*/

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
		if solution.Token != "" {
			coingeckoId := tokenConfig[solution.Token].CoingeckoId
			//tokenPrice := fmt.Sprintf("%f", (*prices)[coingeckoId]["usd"])
			tokenPrice := float64((*prices)[coingeckoId]["usd"])
			var solutionModel models.Solution

			db.GetClient().Raw("SELECT * FROM solution WHERE name = ?", solution.Name).Scan(&solutionModel)
			solutionModel.TokenPriceFloat = tokenPrice
			db.GetClient().Save(&solutionModel)

		}
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

func CalculateFees() {
	res, err := http.Get("https://l2fees.info")
	if err != nil {
		fmt.Println("Error occured while getting page: ", err)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error occured while parsing response body: ", err)
		return
	}

	doc.Find(".jsx-3095944076.item").Each(func(i int, s *goquery.Selection) {

		nameDiv := s.Find(".jsx-569913960.name").First()
		name := nameDiv.Find(".jsx-569913960").First().Text()
		// For each item found, get the title
		fee := s.Find(".amount").First().Text()
		fee2 := s.Find(".amount").First().Next().Text()
		fmt.Println(name, " : ", fee, ", ", fee2)
	})
	fmt.Println("scraping done")
}

/*
func CalculateTps() {
	gtk.Init(nil)
	go func() {
		runtime.LockOSThread()
		gtk.Main()
	}()

	ctx := webloop.New()
	view := ctx.NewView()
	defer view.Close()
	view.Open("https://ethtps.info")
	err := view.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load URL: %s", err)
		os.Exit(1)
	}
	res, err := view.EvaluateJavaScript("document.title")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run JavaScript: %s", err)
		os.Exit(1)
	}

	fmt.Println(res)

	fmt.Println("scraping done")
}
*/
