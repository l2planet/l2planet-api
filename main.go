package main

import (
	"github.com/l2planet/l2planet-api/src/controllerloops"
)

func main() {
	/*
		dat, _ := os.ReadFile("./config/arbitrum.yaml")
		var chainConfig ChainConfig
		if err := yaml.Unmarshal(dat, &chainConfig); err != nil {
			fmt.Println(err)
			return
		}

		for _, v := range chainConfig.Bridges {
			res, _ := alchemy.GetTokenBalancesOfAnAddress(v.Address)
			fmt.Println(v.Address)
			for k, v := range res {
				fmt.Println(k, v)
			}
		}
	*/
	controllerloops.CalculateTvl()
}
