package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Chain struct {
	gorm.Model
	StringID    string         `json:"string_id"`
	Icon        string         `json:"icon" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description" gorm:"not null"`
	EvmID       int            `jspn:"evm_id"`
	Solutions   pq.StringArray `json:"layer2s" gorm:"type:text[]"`
}

func (Chain) TableName() string { return "chains" }

type Solution struct {
	gorm.Model
	ChainID         string         `json:"chain_id"`
	StringID        string         `json:"string_id"`
	Categories      pq.StringArray `gorm:"type:text[]" json:"categories"`
	Name            string         `gorm:"unique;not null" json:"name"`
	Icon            string         `json:"icon"`
	Website         string         `json:"website"`
	Twitter         string         `json:"twitter"`
	Github          string         `json:"github"`
	Videos          pq.StringArray `gorm:"type:text[]" json:"video"`
	CoinGecko       string         `json:"gecko"`
	Investors       pq.StringArray `gorm:"type:text[]" json:"investors"`
	Description     string         `json:"description"`
	Token           string         `json:"token"`
	TokenPriceFloat float64        `json:"token_price"`
	Bridges         []Bridge
	Projects        pq.StringArray `gorm:"type:text[]" json:"projects"`
	EvmID           string         `json:"evm_id"`
}

func (Solution) TableName() string { return "solution" }

type Project struct {
	gorm.Model
	StringID    string         `json:"string_id"`
	Icon        string         `json:"icon"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Website     string         `json:"website"`
	Twitter     string         `json:"twitter"`
	Github      string         `json:"github"`
	Video       string         `json:"video"`
	Investors   pq.StringArray `json:"investors" gorm:"type:text[]"`
	SolutionID  uint           `gorm:"not null"`
	Layer2IDs   pq.StringArray `json:"l2_ids" gorm:"type:text[]"`
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
	ContractAdress  string         `gorm:"unique;not null" json:"contract_address"`
	SupportedTokens pq.StringArray `gorm:"type:text[]" json:"tokens"`
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
	UserName string `json:"username"`
	Text     string `json:"text"`
	UsersID  uint   `json:"user_id"`
}

func (Newsletter) TableName() string { return "newsletters" }

type Users struct {
	gorm.Model
	Newsletters []Newsletter
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func (Users) TableName() string { return "users" }
