package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	Symbol          string  `gorm:"unique;not null"`
	CoinGeckoId     string  `gorm:"unique;not null"`
	Decimals        int     `gorm:"not null"`
	ContractAddress string  `gorm:"unique;not null"`
	Prices          []Price `gorm:"foreignKey:Symbol;references:Symbol"`
}

func (Token) TableName() string { return "tokens" }

type Price struct {
	gorm.Model
	Symbol    string
	Value     float32
	Timestamp time.Time
}

func (Price) TableName() string { return "prices" }

type Solution struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Description string
	Tokens      pq.StringArray `gorm:"type:text[]"`
	Bridges     []Bridge
}

func (Solution) TableName() string { return "solution" }

type Bridge struct {
	gorm.Model
	Native          bool //`gorm:"not null"`
	SolutionID      uint
	ContractAdress  string         `gorm:"unique;not null"`
	SupportedTokens pq.StringArray `gorm:"type:text[]"`
	Balances        []Balance
	Tvls            []Tvl
}

func (Bridge) TableName() string { return "bridges" }

type Balance struct {
	gorm.Model
	Symbol    string    `gorm:"not null"`
	Value     string    `gorm:"not null"`
	Timestamp time.Time `gorm:"not null"`
	BridgeID  uint      `gorm:"not null"`
}

func (Balance) TableName() string { return "balances" }

type Tvl struct {
	gorm.Model
	Value     string `gorm:"not null"`
	Timestamp time.Time
	BridgeID  uint
}

func (Tvl) TableName() string { return "tvls" }

type Blog struct {
	gorm.Model
	UserName  string
	PublicKey string
}

func (Blog) TableName() string { return "blogs" }
