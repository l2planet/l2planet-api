package controllerloops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/models"
)

var (
	rpcToName = map[string]string{
		"https://avax.boba.network":                            "boba_avax",
		"https://mainnet.boba.network":                         "boba_network",
		"https://bobaopera.boba.network":                       "bobaopera",
		"https://bobabeam.boba.network":                        "bobabeam",
		"https://replica.bnb.boba.network":                     "bobabnb",
		"https://rpc-mainnet-cardano-evm.c1.milkomeda.com":     "milkomeda_c1",
		"https://rpc-mainnet-algorand-rollup.a1.milkomeda.com": "milkomeda_a1",
	}
)

type ExplorerResult struct {
	Jsonrpc        string `json:"jsonrpc"`
	ID             int    `json:"id"`
	GetBlockResult struct {
		BaseFeePerGas    string        `json:"baseFeePerGas"`
		Difficulty       string        `json:"difficulty"`
		ExtraData        string        `json:"extraData"`
		GasLimit         string        `json:"gasLimit"`
		GasUsed          string        `json:"gasUsed"`
		Hash             string        `json:"hash"`
		L1BlockNumber    string        `json:"l1BlockNumber"`
		LogsBloom        string        `json:"logsBloom"`
		Miner            string        `json:"miner"`
		MixHash          string        `json:"mixHash"`
		Nonce            string        `json:"nonce"`
		Number           string        `json:"number"`
		ParentHash       string        `json:"parentHash"`
		ReceiptsRoot     string        `json:"receiptsRoot"`
		SendCount        string        `json:"sendCount"`
		SendRoot         string        `json:"sendRoot"`
		Sha3Uncles       string        `json:"sha3Uncles"`
		Size             string        `json:"size"`
		StateRoot        string        `json:"stateRoot"`
		Timestamp        string        `json:"timestamp"`
		TotalDifficulty  string        `json:"totalDifficulty"`
		Transactions     []string      `json:"transactions"`
		TransactionsRoot string        `json:"transactionsRoot"`
		Uncles           []interface{} `json:"uncles"`
	} `json:"result"`
}

func CalculateTpsandGpsFromBlockchain(rpcUrl string, timewindow time.Duration) (float64, float64, error) {
	timeWindowSeconds := uint64(timewindow.Seconds())
	client, _ := ethclient.Dial(rpcUrl)
	latestBlock, err := client.BlockNumber(context.TODO())
	if err != nil {
		return 0, 0, err
	}

	startBlock, err := client.BlockByNumber(context.TODO(), big.NewInt(int64(latestBlock)))
	if err != nil {
		return 0, 0, err
	}

	startTime := startBlock.Header().Time
	finishTime := uint64(0)
	diff := uint64(0)
	txCounter := len(startBlock.Transactions())
	gasCounter := int(startBlock.GasUsed())

	i := 1
	for {
		block, err := client.BlockByNumber(context.TODO(), big.NewInt(int64(latestBlock-uint64(i))))
		if err != nil {
			continue
		}
		txCounter += len(block.Transactions())
		gasCounter += int(block.GasUsed())
		finishTime = block.Header().Time
		diff = startTime - finishTime
		if diff >= timeWindowSeconds {
			break
		}
		i++
	}
	return float64(txCounter) / float64(diff), float64(gasCounter) / float64(timewindow), nil
}

func CalculateMissingTps() {
	var sol models.Solution
	tps, _, err := CalculateTpsandGpsFromExplorerArbiNova()
	if err != nil {
		fmt.Print(err)
	}

	if tx := db.GetClient().Raw("SELECT * FROM solution WHERE string_id = ?", "arbitrum_nova").Scan(&sol); tx.Error == nil && tx.RowsAffected != 0 {
		sol.Tps = fmt.Sprintf("%f", tps)
		db.GetClient().Save(&sol)
	}

	rpcList := []string{"https://avax.boba.network", "https://mainnet.boba.network", "https://bobaopera.boba.network", "https://bobabeam.boba.network", "https://replica.bnb.boba.network", "https://rpc-mainnet-cardano-evm.c1.milkomeda.com", "https://rpc-mainnet-algorand-rollup.a1.milkomeda.com"}
	for _, rpcUrl := range rpcList {
		tps, _, err := CalculateTpsandGpsFromBlockchain(rpcUrl, 60*time.Second)
		if err != nil {
			fmt.Print(err)
		}
		if tx := db.GetClient().Raw("SELECT * FROM solution WHERE string_id = ?", rpcToName[rpcUrl]).Scan(&sol); tx.Error == nil && tx.RowsAffected != 0 {
			sol.Tps = fmt.Sprintf("%f", tps)
			db.GetClient().Save(&sol)
		}
	}
}

func CalculateTpsandGpsFromExplorerArbiNova() (float64, float64, error) {
	//apiKey := "F55JAZG4RBQ2RYTVDEM1WAMFT647EINUBI"
	//apiKey := "8XHABKYCMZ3ZAKZRZE7F4AX49ZSN1VEY5W"
	getBlockFormat := "https://api-nova.arbiscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%x&boolean=false&apikey=F55JAZG4RBQ2RYTVDEM1WAMFT647EINUBI"
	getBlockNumUrl := "https://api-nova.arbiscan.io/api?module=proxy&action=eth_blockNumber&apikey=F55JAZG4RBQ2RYTVDEM1WAMFT647EINUBI"
	resp, err := http.Get(getBlockNumUrl)
	if err != nil {
		return 0, 0, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var blockNumMap map[string]interface{}
	if err := json.Unmarshal(respBody, &blockNumMap); err != nil {
		return 0, 0, err
	}

	hexStartBlockNum := blockNumMap["result"].(string)
	decimalStartBlockNum, err := strconv.ParseInt(hexStartBlockNum[2:], 16, 64)
	if err != nil {
		return 0, 0, err
	}
	startTime := int64(0)
	finishTime := int64(0)
	totalGasUsed := int64(0)
	totalTxCount := int64(0)
	for i := 0; i < 10; i++ {
		var explorerResult ExplorerResult
		decimalBlockNum := decimalStartBlockNum - int64(i)
		getBlockUrl := fmt.Sprintf(getBlockFormat, decimalBlockNum)
		resp, err = http.Get(getBlockUrl)
		if err != nil {
			return 0, 0, err
		}

		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return 0, 0, err
		}

		if err := json.Unmarshal(respBody, &explorerResult); err != nil {
			return 0, 0, err
		}

		blockGas, err := strconv.ParseInt(explorerResult.GetBlockResult.GasUsed[2:], 16, 64)
		if err != nil {
			return 0, 0, err
		}

		blockTs, err := strconv.ParseInt(explorerResult.GetBlockResult.Timestamp[2:], 16, 64)
		if err != nil {
			return 0, 0, err
		}
		if i == 0 {
			startTime = blockTs
		}
		finishTime = blockTs
		blockTxCount := int64(len(explorerResult.GetBlockResult.Transactions))
		totalGasUsed += blockGas
		totalTxCount += blockTxCount
		time.Sleep(500 * time.Millisecond)
	}
	timePassed := startTime - finishTime
	return float64(totalTxCount) / float64(timePassed), float64(totalGasUsed) / float64(timePassed), nil
}
