package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/models"
)

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
	}

	if err := db.GetClient().Create(&solution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

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
	chains, err := db.GetClient().GetAllChains()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	chainsMap := make(map[string]models.Chain)
	for _, chain := range chains {
		chainsMap[chain.Name] = chain
	}

	solutionsWithTvl, err := db.GetClient().GetAllSolutionsWithTvl()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	solutionsMap := make(map[string]db.SolutionWithTvl)
	for _, solution := range solutionsWithTvl {
		solutionsMap[solution.Name] = solution
	}

	projects, err := db.GetClient().GetAllProjects()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	projectsMap := make(map[string]models.Project)
	for _, project := range projects {
		projectsMap[project.Name] = project
	}

	newsletter, err := db.GetClient().GetLatestNewsletter()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	/*
		chainsByte, err := json.Marshal(chains)
		if err != nil {

			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		solutionsWithTvlbyte, err := json.Marshal(solutionsWithTvl)
		if err != nil {

			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		projectByte, err := json.Marshal(projects)
		if err != nil {

			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		newsletterByte, err := json.Marshal(newsletter)
		if err != nil {

			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		c.JSON(
			http.StatusOK, gin.H{
				"chains":            string(chainsByte),
				"solutions":         string(solutionsWithTvlbyte),
				"projects":          string(projectByte),
				"latest_newsletter": string(newsletterByte),
			})
	*/
	c.JSON(
		http.StatusOK, gin.H{
			"chains":            chainsMap,
			"solutions":         solutionsMap,
			"projects":          projectsMap,
			"latest_newsletter": newsletter,
		})
}
