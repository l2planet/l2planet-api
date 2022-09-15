package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Chain struct {
	gorm.Model
	Icon        string `json:"icon"`
	Name        string `json:"name"`
	Description string `json:"description"`
	EvmID       int
	Solutions   pq.StringArray `json:"solutions" gorm:"type:text[]"`
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
	EvmID       int
	//ChainID     uint
}

func (Solution) TableName() string { return "solution" }

type Project struct {
	Icon        string         `json:"icon"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Website     string         `json:"website"`
	Twitter     string         `json:"twitter"`
	Github      string         `json:"github"`
	Video       string         `json:"video"`
	Investors   pq.StringArray `json:"investors" gorm:"type:text[]"`
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
