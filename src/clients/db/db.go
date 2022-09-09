package db

import (
	"os"

	"github.com/l2planet/l2planet-api/src/models"
	"github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BridgeConfig struct {
	Address string   `yaml:"address"`
	Tokens  []string `yaml:"tokens"`
}

type SolutionConfig struct {
	Bridges     []BridgeConfig `yaml:"bridges"`
	Name        string         `yaml:"name"`
	Tokens      []string       `yaml:"tokens"`
	Description string         `yaml:"description"`
}

type SolutionWithTvl struct {
	Name        string
	Icon        string
	Website     string
	Twitter     string
	Github      string
	Video       string
	Investors   pq.StringArray `gorm:"type:text[]"`
	Description string
	Tokens      pq.StringArray `gorm:"type:text[]"`
	Projects    pq.StringArray `gorm:"type:text[]"`
	SolutionID  string
	TvlValue    float64
}

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

func (c *Client) GetAllChains() ([]models.Chain, error) {
	var chains []models.Chain

	if err := c.DB.Find(&chains).Error; err != nil {
		return nil, err
	}

	return chains, nil
}

func (c *Client) GetAllProjects() ([]models.Project, error) {
	var projects []models.Project

	if err := c.DB.Find(&projects).Error; err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *Client) GetAllSolutionsWithTvl() ([]SolutionWithTvl, error) {
	var solutionWithTvls []SolutionWithTvl
	err := c.Raw("SELECT * FROM solution INNER JOIN (SELECT DISTINCT ON(bridges.solution_id) bridges.solution_id,sum(tvls.value) as tvl_value,tvls.timestamp FROM bridges INNER JOIN tvls on bridges.id = tvls.bridge_id GROUP BY solution_id,tvls.timestamp ORDER BY bridges.solution_id,tvls.timestamp DESC) as bridgetvl ON solution.id = bridgetvl.solution_id").Scan(&solutionWithTvls).Error
	if err != nil {
		return nil, err
	}

	return solutionWithTvls, nil
}

func (c *Client) GetSolutionConfig() ([]models.Solution, error) {
	var solutions []models.Solution
	if err := c.DB.Model(&models.Solution{}).Preload("Bridges").Find(&solutions).Error; err != nil {
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
		var chainConfig SolutionConfig
		if err := yaml.Unmarshal(dat, &chainConfig); err != nil {
			panic(err)
		}

		solution := &models.Solution{
			Name:        chainConfig.Name,
			Tokens:      chainConfig.Tokens,
			Description: chainConfig.Description,
		}

		c.Create(solution)

		for _, bridge := range chainConfig.Bridges {
			bridgeModel := &models.Bridge{
				SolutionID:      solution.Model.ID,
				ContractAdress:  bridge.Address,
				SupportedTokens: bridge.Tokens,
			}
			c.Create(bridgeModel)
		}
	}
	return nil
}
