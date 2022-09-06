package db

import (
	"os"

	"github.com/l2planet/l2planet-api/src/models"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BridgeConfig struct {
	Address string   `yaml:"address"`
	Tokens  []string `yaml:"tokens"`
}

type ChainConfig struct {
	Bridges     []BridgeConfig `yaml:"bridges"`
	Name        string         `yaml:"name"`
	Tokens      []string       `yaml:"tokens"`
	Description string         `yaml:"description"`
}

var db *gorm.DB

var client *Client

type Client struct {
	*gorm.DB
}

func GetClient() *Client {
	if client == nil {
		dbInstance, err := gorm.Open(postgres.New(postgres.Config{
			DSN:                  "host=localhost user=postgres password=123456789 dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai",
			PreferSimpleProtocol: true, // disables implicit prepared statement usage
		}), &gorm.Config{})
		if err != nil {
			panic(err.Error())
		}
		client = &Client{
			DB: dbInstance,
		}
	}
	return client
}

func GetDbClient() *gorm.DB {
	var err error
	if db == nil {
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  "host=localhost user=postgres password=123456789 dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai",
			PreferSimpleProtocol: true, // disables implicit prepared statement usage
		}), &gorm.Config{})
		if err != nil {
			panic(err.Error())
		}
	}
	return db

}

func (c *Client) GetSolutionConfig() ([]models.Solution, error) {
	var solutions []models.Solution
	//if err := db.Joins("INNER JOIN bridges ON bridges.solution_id = solutions.id").Find(&solutions).Error; err != nil {
	if err := db.Model(&models.Solution{}).Preload("Bridges").Find(&solutions).Error; err != nil {
		//if err := db.Joins("bridges").Find(&solutions).Error; err != nil {
		return nil, err
	}

	return solutions, nil
}

func (c *Client) SyncDb() error {
	files, err := os.ReadDir("./config/chains")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		dat, _ := os.ReadFile("./config/chains/" + file.Name())
		var chainConfig ChainConfig
		if err := yaml.Unmarshal(dat, &chainConfig); err != nil {
			panic(err)
		}

		l2 := &models.Solution{
			Name:        chainConfig.Name,
			Tokens:      chainConfig.Tokens,
			Description: chainConfig.Description,
		}

		c.Create(l2)

		for _, bridge := range chainConfig.Bridges {
			bridgeModel := &models.Bridge{
				SolutionID:      l2.Model.ID,
				ContractAdress:  bridge.Address,
				SupportedTokens: bridge.Tokens,
			}
			c.Create(bridgeModel)
		}
	}
	return nil
}
