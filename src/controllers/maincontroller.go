package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/clients/redis"
	"github.com/l2planet/l2planet-api/src/consts"
	"github.com/l2planet/l2planet-api/src/models"
)

func Register(c *gin.Context) {
	var user models.Users
	if err := c.BindJSON(&user); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	user.Password = string(pass)
	if err := db.GetClient().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func NewNewsletter(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	var newsletter models.Newsletter
	if err := c.BindJSON(&newsletter); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	newsletter.UserName = claims[consts.IdentityName].(string)
	newsletter.UsersID = uint(claims[consts.IdentityKey].(float64))

	if err := db.GetClient().Create(&newsletter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func NewChain(c *gin.Context) {
	log.Printf("New Chain")
	var chain models.Chain
	if err := c.BindJSON(&chain); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().Create(&chain).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func PatchChain(c *gin.Context) {
	log.Printf("Patch Chain")
	var chain models.Chain
	var chainQuery models.Chain
	if err := c.BindJSON(&chain); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	if err := db.GetClient().First(&chainQuery, "string_id = ?", chain.StringID).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	chain.ID = chainQuery.ID

	if err := db.GetClient().Save(&chain).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func NewSolution(c *gin.Context) {
	log.Printf("New Solution")
	var solution models.Solution
	if err := c.BindJSON(&solution); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().Create(&solution).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	var chain models.Chain
	db.GetClient().Raw("SELECT * FROM chains WHERE string_id = ?", solution.ChainID).Scan(&chain)
	chain.Solutions = append(chain.Solutions, solution.StringID)
	db.GetClient().Save(&chain)

	c.JSON(http.StatusOK, nil)
}

func PatchSolution(c *gin.Context) {
	log.Printf("Patch Solution")
	var solution models.Solution
	var solutionQuery models.Solution
	if err := c.BindJSON(&solution); err != nil {
		fmt.Println("patch solution error")
                fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().First(&solutionQuery, "string_id = ?", solution.StringID).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	solution.ID = solutionQuery.ID

	for i, bridge := range solution.Bridges {
		if bridge.ContractAdress == "" {
			continue
		}
		bridgeQuery := models.Bridge{}
		if err := db.GetClient().First(&bridgeQuery, "contract_adress = ?", bridge.ContractAdress).Error; err != nil && err != gorm.ErrRecordNotFound {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, nil)
			return
		} else if err == gorm.ErrRecordNotFound {
			continue
		}
		solution.Bridges[i].ID = bridgeQuery.ID

	}

	if err := db.GetClient().Save(&solution).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func NewProject(c *gin.Context) {
	log.Printf("New Project")
	body, _ := ioutil.ReadAll(c.Request.Body)
	println(string(body))
	var project models.Project
	if err := c.BindJSON(&project); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	var solution models.Solution
	for _, l2 := range project.Layer2IDs {
		db.GetClient().Raw("SELECT * FROM solution WHERE string_id = ?", l2).Scan(&solution)
		solution.Projects = append(solution.Projects, project.StringID)
		db.GetClient().Save(&solution)
	}

	c.JSON(http.StatusOK, nil)
}

func PatchProject(c *gin.Context) {
	log.Printf("Patch Project")
	body, _ := ioutil.ReadAll(c.Request.Body)
	println(string(body))
	var project models.Project
	var projectQuery models.Project
	if err := c.BindJSON(&project); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().First(&projectQuery, "string_id = ?", project.StringID).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	project.ID = projectQuery.ID

	if err := db.GetClient().Save(&project).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func Raw(c *gin.Context) {
	responseMap := make(map[string]interface{}, 0)
	cacheRes, err := redis.GetClient().Get(context.TODO(), "raw").Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json", []byte(cacheRes))
		return
	}

	chains, err := db.GetClient().GetAllChains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	chainsMap := make(map[string]models.Chain)
	for _, chain := range chains {
		chainsMap[chain.StringID] = chain
	}
	responseMap["chains"] = chainsMap

	solutions, err := db.GetClient().GetAllSolutionsWithBridges()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	solutionsMap := make(map[string]models.Solution)
	for _, solution := range solutions {
		solutionsMap[solution.StringID] = solution
	}
	responseMap["layer2s"] = solutionsMap

	projects, err := db.GetClient().GetAllProjects()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	projectsMap := make(map[string]models.Project)
	for _, project := range projects {
		projectsMap[project.StringID] = project
	}
	responseMap["projects"] = projectsMap

	newsletter, err := db.GetClient().GetAllNewsletters()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	responseMap["newsletters"] = newsletter

	responseBody, err := json.Marshal(responseMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	redis.GetClient().Set(context.TODO(), "raw", string(responseBody), 5*time.Minute).Result()

	c.Data(http.StatusOK, "application/json", responseBody)
}

func RawLayer2(c *gin.Context) {
	cacheRes, err := redis.GetClient().Get(context.TODO(), "raw_layer2").Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json", []byte(cacheRes))
		return
	}

	solutions, err := db.GetClient().GetAllSolutionsWithBridges()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	solutionsMap := make(map[string]models.Solution)
	for _, solution := range solutions {
		solutionsMap[solution.StringID] = solution
	}

	responseBody, err := json.Marshal(solutionsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	redis.GetClient().Set(context.TODO(), "raw_layer2", string(responseBody), 5*time.Minute).Result()

	c.Data(http.StatusOK, "application/json", responseBody)
}

func RawNewsletter(c *gin.Context) {
	cacheRes, err := redis.GetClient().Get(context.TODO(), "raw_newsletter").Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json", []byte(cacheRes))
		return
	}

	newsletter, err := db.GetClient().GetAllNewsletters()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	responseBody, err := json.Marshal(newsletter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	redis.GetClient().Set(context.TODO(), "raw_newsletter", string(responseBody), 5*time.Minute).Result()

	c.Data(http.StatusOK, "application/json", responseBody)
}

func Info(c *gin.Context) {
	responseMap := make(map[string]interface{}, 0)
	infoResponse := make([]db.InfoResponse, 0)
	cacheRes, err := redis.GetClient().Get(context.TODO(), "info").Result()
	if err == nil {
		c.Data(http.StatusOK, "application/json", []byte(cacheRes))
		return
	}

	chains, err := db.GetClient().GetAllChains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	chainsMap := make(map[string]models.Chain)
	for _, chain := range chains {
		chainsMap[chain.StringID] = chain
	}
	responseMap["chains"] = chainsMap

	solutionsWithTvl, err := db.GetClient().GetAllSolutionsWithTvl()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	/*
		daily, err := db.GetClient().GetAllTvlsWithLength("hour", "HH24:00", 24)
		if err != nil {

			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	*/
	yearly, err := db.GetClient().GetAllTvlsWithLength("day", "yyyy-mm--dd", 365)
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	for _, sol := range solutionsWithTvl {
		//length := len(yearly[sol.Name])
		infoResponse = append(infoResponse, db.InfoResponse{
			SolutionWithTvl: sol,
			Tvls:            yearly[sol.Name],
			/*db.Tvl{
				Daily:     daily[sol.Name],
				Weekly:    yearly[sol.Name][positiveOrZero(length-7):],
				Monthly:   yearly[sol.Name][positiveOrZero(length-30):],
				Quarterly: yearly[sol.Name][positiveOrZero(length-90):],
				Yearly:    yearly[sol.Name],
			},*/
		})
	}

	solutionsMap := make(map[string]db.InfoResponse)
	for _, solution := range infoResponse {
		solutionsMap[solution.StringID] = solution
	}
	responseMap["layer2s"] = solutionsMap

	projects, err := db.GetClient().GetAllProjects()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	projectsMap := make(map[string]models.Project)
	for _, project := range projects {
		projectsMap[project.StringID] = project
	}
	responseMap["projects"] = projectsMap

	newsletter, err := db.GetClient().GetLatestNewsletter()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	responseMap["latest_newsletter"] = newsletter

	responseBody, err := json.Marshal(responseMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	redis.GetClient().Set(context.TODO(), "info", string(responseBody), 5*time.Minute).Result()
	c.Data(http.StatusOK, "application/json", responseBody)
}

func positiveOrZero(num int) int {
	if num < 0 {
		return 0
	}
	return num
}
