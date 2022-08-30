package coingecko

import (
	"net/http"
	"time"

	cg "github.com/superoo7/go-gecko/v3"
)

type Client struct {
	*cg.Client
}

func NewClient() *Client {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	coinGeckoClient := cg.NewClient(httpClient)
	return &Client{
		Client: coinGeckoClient,
	}
}

func (client *Client) GetPrice(id string) (float32, error) {
	price, err := client.Client.SimpleSinglePrice(id, "usd")
	if err != nil {
		return 0.00, err
	}
	return price.MarketPrice, nil
}

func (client *Client) GetPrices(ids []string) (*map[string]map[string]float32, error) {
	prices, err := client.Client.SimplePrice(ids, []string{"usd"})
	if err != nil {
		return nil, err
	}
	return prices, nil
}
