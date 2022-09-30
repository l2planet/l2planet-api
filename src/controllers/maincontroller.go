package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/clients/redis"
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
	var newsletter models.Newsletter
	if err := c.BindJSON(&newsletter); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().Create(&newsletter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func NewChain(c *gin.Context) {
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

func NewSolution(c *gin.Context) {
	var solution models.Solution
	if err := c.BindJSON(&solution); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().Create(&solution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	var chain models.Chain
	db.GetClient().Raw("SELECT * FROM chains WHERE string_id = ?", solution.ChainID).Scan(&chain)
	chain.Solutions = append(chain.Solutions, solution.StringID)
	db.GetClient().Save(&chain)

	c.JSON(http.StatusOK, nil)
}

func NewProject(c *gin.Context) {
	var project models.Project
	if err := c.BindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if err := db.GetClient().Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func Info(c *gin.Context) {
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

	solutionsWithTvl, err := db.GetClient().GetAllSolutionsWithTvl()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	daily, err := db.GetClient().GetAllTvlsWithLength("hour", 24)
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	dailyMap := make(map[string][]string)
	for _, v := range daily {
		dailyMap[v.Name] = v.Historical
	}

	yearly, err := db.GetClient().GetAllTvlsWithLength("day", 365)
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	yearlyMap := make(map[string][]string)
	for _, v := range yearly {
		yearlyMap[v.Name] = v.Historical
	}

	for i, sol := range solutionsWithTvl {

		solutionsWithTvl[i].Tvls = db.HistoricalTvl{
			Daily:  dailyMap[sol.Name],
			Yearly: yearlyMap[sol.Name],
		}
	}

	solutionsMap := make(map[string]db.SolutionWithTvl)
	for _, solution := range solutionsWithTvl {
		solutionsMap[solution.StringID] = solution
	}

	projects, err := db.GetClient().GetAllProjects()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	projectsMap := make(map[string]models.Project)
	for _, project := range projects {
		projectsMap[project.StringID] = project
	}

	newsletter, err := db.GetClient().GetLatestNewsletter()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	responseMap := make(map[string]interface{}, 0)
	responseMap["chains"] = chainsMap
	responseMap["layer2s"] = solutionsMap
	responseMap["projects"] = projectsMap
	responseMap["latest_newsletter"] = newsletter

	responseBody, err := json.Marshal(responseMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	redis.GetClient().Set(context.TODO(), "info", string(responseBody), 5*time.Minute).Result()

	//c.PureJSON(http.StatusOK, string(responseBody))
	c.Data(http.StatusOK, "application/json", responseBody)
	//c.PureJSON()
	//c.String(http.StatusOK, "%s", string(responseBody))
}
