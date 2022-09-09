package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/models"
)

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
	return
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
	return
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
	return
}

func Info(c *gin.Context) {
	chains, err := db.GetClient().GetAllChains()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	solutionsWithTvl, err := db.GetClient().GetAllSolutionsWithTvl()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	projects, err := db.GetClient().GetAllProjects()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
		return
	}

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

	c.JSON(
		http.StatusOK, gin.H{
			"chains":    string(chainsByte),
			"solutions": string(solutionsWithTvlbyte),
			"projects":  string(projectByte),
		})
}
