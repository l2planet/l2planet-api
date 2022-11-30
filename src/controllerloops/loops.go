package controllerloops

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/l2planet/l2planet-api/src/clients/coingecko"
	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/clients/ethereum"
	"github.com/l2planet/l2planet-api/src/consts"
	"github.com/l2planet/l2planet-api/src/models"
)

const (
	localDir = "./config/"
)

var (
	defiLlamaMap = map[string]bool{
		"Milkomeda": true,
	}
	feeNameConverter = map[string]string{
		"Loopring":       "Loopring",
		"ZKSync":         "zkSync",
		"Arbitrum One":   "Arbitrum One",
		"Metis Network":  "Metis Andromeda",
		"Boba Network":   "Boba Network",
		"Aztec Network":  "Aztec",
		"Optimism":       "Optimism",
		"Polygon Hermez": "Polygon Hermez",
	}
	tpsNameConverter = map[string]string{
		"Loopring":       "Loopring",
		"ZKSync":         "zkSync",
		"Arbitrum One":   "Arbitrum One",
		"Metis":          "Metis Andromeda",
		"Boba Network":   "Boba Network",
		"Aztec":          "Aztec",
		"Optimism":       "Optimism",
		"Polygon Hermez": "Polygon Hermez",
		"Starknet":       "StarkNet",
		"Arbitrum Nova":  "Arbitrum Nova",
		"Immutable X":    "Immutable X",
		"Sorare":         "Sorare",
		"ZKSwap":         "ZKSwap 1.0",
		"ZKSpace":        "ZKSpace",
		"DeversiFi":      "rhino.fi",
		"OMG Network":    "OMG Network",
		"Fuel":           "Fuel v1",
		"dYdX":           "dYdX",
		"Gluon":          "Gluon",
		"Layer2.Finance": "Layer2.Finance",
	}
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

type DefiLlamaResponse struct {
	GeckoId string  `json:"gecko_id"`
	Tvl     float64 `json:"tvl"`
	Symbol  string  `json:"tokenSymbol"`
	CmcId   string  `json:"cmcId"`
	Name    string  `json:"name"`
}

type TvlAdapter interface {
	GetTokenConfig() (map[string]TokenConfig, []string, error)
	CalculateTvl(ts time.Time) error
}

type EthereumAdapter struct {
	rpcEndpoint string
	client      *ethereum.Client
}

func getTokenConfig(chainId string) (map[string]TokenConfig, []string, error) {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = localDir
	}
	dat, _ := os.ReadFile(configDir + fmt.Sprintf("tokens/%s/tokens.json", chainId))
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

func CalculateTvlAvalanche(ts time.Time) error {
	tokenConfigs, cgSymbolList, _ := getTokenConfig("avalanche")
	for _, token := range tokenConfigs {
		fmt.Println(token.Name, " ", token.Address)
	}
	solutionConfigs, _ := db.GetClient().GetSolutionConfig("avalanche")
	for _, sol := range solutionConfigs {
		fmt.Println(sol.Name, " ")
		for _, bridge := range sol.Bridges {
			fmt.Print(bridge.ID, " ", bridge.ContractAdress, " ")
		}
		fmt.Println()
	}
	avalancheUrl := consts.AvalancheClientUrl
	client := ethereum.NewClient(avalancheUrl)
	err := calculateTvlEvm(client, tokenConfigs, cgSymbolList, ts, solutionConfigs)
	return err
}

func CalculateTvl(ts time.Time) error {
	tokenConfigs, cgSymbolList, _ := getTokenConfig("ethereum")
	solutionConfigs, _ := db.GetClient().GetSolutionConfig("ethereum")
	ethUrl := os.Getenv("ETH_URL")
	if ethUrl == "" {
		ethUrl = consts.EthClientUrl
	}
	client := ethereum.NewClient(ethUrl)
	err := calculateTvlEvm(client, tokenConfigs, cgSymbolList, ts, solutionConfigs)
	return err
}

// TODO: instead of querying blockchain one by one, use multicall
func calculateTvlEvm(client *ethereum.Client, tokenConfig map[string]TokenConfig, cgSymbolList []string, ts time.Time, solutionConfigs []models.Solution) error {
	coinGeckoClient := coingecko.NewClient()

	prices, _ := coinGeckoClient.GetPrices(cgSymbolList)
	tx := db.GetClient().DB.Begin()
	for _, solution := range solutionConfigs {
		if solution.Token != "" {
			coingeckoId := tokenConfig[solution.Token].CoingeckoId
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
					balance, err := getBalance(client, bridge.ContractAdress, tokenConfig[name].Address, tokenConfig[name].Decimals)
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
					balance, err := getBalance(client, bridge.ContractAdress, tokenConfig[tokenName].Address, tokenConfig[tokenName].Decimals)
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

func CalculateFees() error {
	res, err := http.Get("https://l2fees.info")
	if err != nil {
		fmt.Println("Error occured while getting page: ", err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error occured while parsing response body: ", err)
		return err
	}

	doc.Find(".jsx-3095944076.item").Each(func(i int, s *goquery.Selection) {
		nameDiv := s.Find(".jsx-569913960.name").First()
		name := nameDiv.Find(".jsx-569913960").First().Text()
		// For each item found, get the title
		send := strings.ReplaceAll(s.Find(".amount").First().Text(), "$", "")
		swap := strings.ReplaceAll(s.Find(".amount").First().Next().Text(), "$", "")
		var solution models.Solution
		transformedName := feeNameConverter[name]
		if transformedName != "" {
			if tx := db.GetClient().Raw("SELECT * FROM solution WHERE name = ?", transformedName).Scan(&solution); tx.Error == nil && tx.RowsAffected != 0 {
				solution.Send = send
				solution.Swap = swap
				db.GetClient().Save(&solution)
			}
		}

	})
	return nil
}

func CalculateTps() error {
	var TpsMap map[string][]db.Tps

	res, err := http.Get("https://api.ethtps.info/API/TPS/Get?Provider=All&Network=Mainnet&interval=OneMonth")
	if err != nil {
		fmt.Println("Error occured while getting page: ", err)
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error occured while reading response body: ", err)
		return err
	}

	if err := json.Unmarshal(body, &TpsMap); err != nil {
		fmt.Println("Could not unmarshall tps data: ", err)
		return err
	}

	for k, v := range TpsMap {
		total := float64(0)
		for _, val := range v {
			total += val.Data[0].Value
		}

		total = total / float64(len(v))
		var sol models.Solution
		transformedName := tpsNameConverter[k]
		if transformedName != "" {
			if tx := db.GetClient().Raw("SELECT * FROM solution WHERE name = ?", transformedName).Scan(&sol); tx.Error == nil && tx.RowsAffected != 0 {
				sol.Tps = fmt.Sprintf("%f", total)
				db.GetClient().Save(&sol)
			}
		}
	}
	return nil
}

func CalculateTvlViaScrape(ts time.Time) error {
	defiLlamaResponses := make([]DefiLlamaResponse, 0)
	res, err := http.Get("https://api.llama.fi/chains")
	if err != nil {
		fmt.Println("could not get tvl via scrape : ", err)
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("could not read body via scrape : ", err)
		return err
	}
	err = json.Unmarshal(body, &defiLlamaResponses)
	if err != nil {
		fmt.Println("could not unmarshall body via scrape : ", err)
		return err
	}

	for _, defiLlamaResponse := range defiLlamaResponses {
		if defiLlamaMap[defiLlamaResponse.Name] {
			if err := db.GetClient().Create(&models.ScrapedTvl{
				Name:      defiLlamaResponse.Name,
				Timestamp: ts,
				Value:     defiLlamaResponse.Tvl,
			}).Error; err != nil {
				fmt.Println("error while inserting tvls of ", defiLlamaResponse.Name)
			}
		}
	}
	return nil
}
