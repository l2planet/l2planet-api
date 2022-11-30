package controllerloops

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

type AvalancheAdapter struct {
	rpcEndpoint string
	client      *ethclient.Client
}
