package ethereum

import (
	"context"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/l2planet/l2planet-api/src/consts"
	"github.com/l2planet/l2planet-api/src/token"
)

type Client struct {
	*ethclient.Client
}

func NewClient() *Client {
	ethUrl := os.Getenv("ETH_URL")
	if ethUrl == "" {
		ethUrl = consts.EthClientUrl
	}
	client, _ := ethclient.Dial(ethUrl)

	return &Client{
		Client: client,
	}
}

func (client *Client) MulticallBalanceAt(address []string) {
	//client.Client.
}

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
