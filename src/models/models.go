package models

import (
	"time"

	"gorm.io/gorm"
)

type TokenModel struct {
	gorm.Model
	Symbol          string
	CoinGeckoId     string
	Decimals        int
	ContractAddress string
}

type PriceModel struct {
	gorm.Model
	Symbol    string
	Price     float32
	Timestamp time.Time
}

type BalanceModel struct {
	gorm.Model
	Symbol    string
	Balance   string
	Timestamp time.Time
}

type TvlModel struct {
	gorm.Model
	Name      string
	Tvl       string
	Timestamp time.Time
}
