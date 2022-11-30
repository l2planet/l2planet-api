package ethereum

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/l2planet/l2planet-api/src/multicall"
	"github.com/l2planet/l2planet-api/src/token"
)

const (
	multicallAddress = "0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696"
)

type Client struct {
	*ethclient.Client
	Abi              abi.ABI
	MulticallAddress common.Address
}

type CallResponse struct {
	Success    bool   `json:"success"`
	ReturnData []byte `json:"returnData"`
}

func NewClient(url string) *Client {
	client, _ := ethclient.Dial(url)

	mcAbi, err := abi.JSON(strings.NewReader(multicall.MulticallABI))
	if err != nil {
		panic(err)
	}

	address := common.HexToAddress(multicallAddress)

	return &Client{
		Client:           client,
		Abi:              mcAbi,
		MulticallAddress: address,
	}
}

/*
	func (client *Client) MulticallBalance(multicalls []multicall.Multicall2Call) {
		var responses []CallResponse
		instance, err := multicall.NewMulticall(client.MulticallAddress, client)
		if err != nil {
			panic(err)
		}
		callData, err := client.Abi.Pack("aggregate", multicalls)
		if err != nil {
			panic(err)
		}

		// Perform multicall
		resp, err := client.Client.CallContract(context.Background(), ethereum.CallMsg{To: &client.MulticallAddress, Data: callData}, nil)
		if err != nil {
			panic(err)
		}
		resp2, err := instance.Aggregate(randomSigner(), multicalls)
		if err != nil {
			panic(err)
		}

		jso, _ := resp2.MarshalJSON()
		fmt.Println(string(jso))
		fmt.Println(hex.EncodeToString(resp))
		unpackedResp, _ := client.Abi.Unpack("aggregate", resp)

			a, err := json.Marshal(unpackedResp[1])
			if err != nil {
				panic(err)
			}
			fmt.Println(a)

		err = json.Unmarshal([]byte(unpackedResp), &responses)
		if err != nil {
			panic(err)
		}

}
*/
func (client *Client) BalanceAt(address string) (*big.Int, error) {
	account := common.HexToAddress(address)
	balance, err := client.Client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return big.NewInt(0), err
	}

	return balance, nil
}

func (client *Client) BalanceOf(ownerAddr, tokenAddr string) (*big.Int, error) {
	tokenAddress := common.HexToAddress(tokenAddr)
	instance, err := token.NewToken(tokenAddress, client)
	if err != nil {
		return big.NewInt(0), err
	}

	address := common.HexToAddress(ownerAddr)
	balance, err := instance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		return big.NewInt(0), err
	}
	return balance, nil
}

func randomSigner() *bind.TransactOpts {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	signer, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
	if err != nil {
		panic(err)
	}

	signer.Context = context.Background()
	signer.GasPrice = big.NewInt(0)

	return signer
}
