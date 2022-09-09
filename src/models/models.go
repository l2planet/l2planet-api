package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Chain struct {
	gorm.Model
	Icon        string
	Name        string
	Description string
	Solutions   pq.StringArray
}

func (Chain) TableName() string { return "chains" }

type Solution struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Icon        string
	Website     string
	Twitter     string
	Github      string
	Video       string
	Investors   pq.StringArray `gorm:"type:text[]"`
	Description string
	Tokens      pq.StringArray `gorm:"type:text[]"`
	Bridges     []Bridge
	Projects    pq.StringArray `gorm:"type:text[]"`
	//ChainID     uint
}

func (Solution) TableName() string { return "solution" }

type Project struct {
	Icon        string
	Name        string
	Description string
	Website     string
	Twitter     string
	Github      string
	Video       string
	Investors   string
	SolutionID  uint
}

func (Project) TableName() string { return "projects" }

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
	Value     float64   `sql:"type:decimal(18,4);"`
	Timestamp time.Time `gorm:"not null"`
	BridgeID  uint      `gorm:"not null"`
}

func (Balance) TableName() string { return "balances" }

type Tvl struct {
	gorm.Model
	Value     float64 `sql:"type:decimal(18,4);"`
	Timestamp time.Time
	BridgeID  uint
}

func (Tvl) TableName() string { return "tvls" }

type Newsletter struct {
	gorm.Model
	UserName  string
	PublicKey string
}

func (Newsletter) TableName() string { return "newsletters" }
