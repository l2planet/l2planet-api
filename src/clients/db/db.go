package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/l2planet/l2planet-api/src/models"
	"github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	databaseUrl = "host=l2planet_db user=postgres password=123456789 dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	localDir    = "./config/"
)

type BridgeConfig struct {
	Address string   `yaml:"address"`
	Tokens  []string `yaml:"tokens"`
}

type SolutionConfig struct {
	Bridges     []BridgeConfig `yaml:"bridges"`
	Name        string         `yaml:"name"`
	Token       string         `yaml:"tokens"`
	Description string         `yaml:"description"`
}

type HistoricalTvl struct {
	Daily  []string `json:"daily"`
	Yearly []string `json:"yearly"`
}

type SolutionWithTvl struct {
	Categories      pq.StringArray `gorm:"type:text[]" json:"categories"`
	ChainID         string         `json:"chain_id"`
	StringID        string         `json:"string_id"`
	ID              uint           `json:"id"`
	Name            string         `json:"name"`
	Icon            string         `json:"icon"`
	Website         string         `json:"website"`
	Twitter         string         `json:"twitter"`
	Github          string         `json:"github"`
	Videos          pq.StringArray `gorm:"type:text[]" json:"videos"`
	Investors       pq.StringArray `gorm:"type:text[]" json:"investors"`
	Description     string         `json:"description"`
	Token           string         `json:"token"`
	TokenPriceFloat float64        `json:"token_price"`
	Projects        pq.StringArray `gorm:"type:text[]" json:"projects"`
	SolutionID      string         `json:"solution_id"`
	TvlValue        float64        `json:"tvl"`
	Tvls            HistoricalTvl  `json:"tvls"`
	CoinGecko       string         `json:"gecko"`
}

type HistoricalTvlModel struct {
	Name       string
	Historical pq.StringArray `gorm:"type:text[]"`
}

var client *Client

type Client struct {
	*gorm.DB
	*pgxpool.Pool
}

func GetClient() *Client {
	if client == nil {
		dbInstance, err := gorm.Open(postgres.New(postgres.Config{
			DSN:                  databaseUrl,
			PreferSimpleProtocol: true, // disables implicit prepared statement usage
		}), &gorm.Config{})
		if err != nil {
			panic(err.Error())
		}
		dbPool, err := pgxpool.Connect(context.Background(), databaseUrl)
		if err != nil {
			panic(err.Error())
		}
		client = &Client{
			DB:   dbInstance,
			Pool: dbPool,
		}
	}
	return client
}

func (c *Client) GetAllChains() ([]models.Chain, error) {
	var chains []models.Chain

	if err := c.DB.Find(&chains).Error; err != nil {
		return nil, err
	}

	/*
		rows, err := c.Pool.Query(context.Background(), "SELECT * FROM chains")
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			chain := models.Chain{}
			err := rows.Scan(chain.ID, chain.CreatedAt, chain.UpdatedAt, chain.DeletedAt, chain.Icon, chain.Name, &chain.Description, &chain.EvmID, &chain.Solutions)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			chains = append(chains, chain)

		}
	*/
	return chains, nil
}

func (c *Client) GetAllProjects() ([]models.Project, error) {
	var projects []models.Project

	if err := c.DB.Find(&projects).Error; err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *Client) GetLatestNewsletter() (models.Newsletter, error) {
	var newsletter models.Newsletter

	if err := c.Raw("SELECT * FROM newsletters ORDER BY created_at DESC LIMIT 1").Scan(&newsletter).Error; err != nil {
		return models.Newsletter{}, err
	}

	return newsletter, nil
}

func (c *Client) GetAllTvlsWithLength(truncateBy string, count int) ([]HistoricalTvlModel, error) {
	var historical []HistoricalTvlModel
	if err := c.Raw("SELECT  sbtwithrow.name,array_agg('{ \"t\": ' || sbtwithrow.timestamp || ', \"v\": ' || sbtwithrow.tvl_value || '}') as historical FROM (SELECT ROW_NUMBER() OVER (PARTITION BY sbt.solution_id ORDER BY sbt.name) AS r,sbt.id,sbt.name,sbt.tvl_value,sbt.timestamp FROM (SELECT DISTINCT ON (date_trunc(?, bridgetvl.timestamp), solution.id) * FROM solution INNER JOIN (SELECT bridges.solution_id,sum(tvls.value) as tvl_value,tvls.timestamp FROM bridges INNER JOIN tvls on bridges.id = tvls.bridge_id GROUP BY solution_id,tvls.timestamp ORDER BY bridges.solution_id,tvls.timestamp DESC) as bridgetvl ON solution.id = bridgetvl.solution_id ORDER BY solution.id, date_trunc(?, bridgetvl.timestamp), bridgetvl.timestamp  DESC) as sbt) as sbtwithrow WHERE sbtwithrow.r <= ? GROUP BY sbtwithrow.name", truncateBy, truncateBy, count).Scan(&historical).Error; err != nil {
		return nil, err
	}
	return historical, nil
}

func (c *Client) GetAllSolutionsWithTvl() ([]SolutionWithTvl, error) {
	var solutionWithTvls []SolutionWithTvl
	if err := c.Raw("SELECT * FROM solution INNER JOIN (SELECT DISTINCT ON(bridges.solution_id) bridges.solution_id,sum(tvls.value) as tvl_value,tvls.timestamp FROM bridges INNER JOIN tvls on bridges.id = tvls.bridge_id GROUP BY solution_id,tvls.timestamp ORDER BY bridges.solution_id,tvls.timestamp DESC) as bridgetvl ON solution.id = bridgetvl.solution_id").Scan(&solutionWithTvls).Error; err != nil {
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
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = localDir
	}

	files, err := os.ReadDir(configDir + "solutions/")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		dat, _ := os.ReadFile(configDir + "solutions/" + file.Name())
		var chainConfig SolutionConfig
		if err := yaml.Unmarshal(dat, &chainConfig); err != nil {
			panic(err)
		}

		solution := &models.Solution{
			Name:        chainConfig.Name,
			Token:       chainConfig.Token,
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

func (c *Client) GetUser(userName string) (*models.Users, error) {
	var user models.Users

	if err := c.Raw("SELECT * FROM users WHERE username = ?", userName).Scan(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
