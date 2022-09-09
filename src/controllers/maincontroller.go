package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/l2planet/l2planet-api/src/clients/db"
)

func Info(c *gin.Context) {
	chains, err := db.GetClient().GetAllChains()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
	}

	solutionsWithTvl, err := db.GetClient().GetAllSolutionsWithTvl()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
	}

	projects, err := db.GetClient().GetAllProjects()
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
	}

	chainsByte, err := json.Marshal(chains)
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
	}

	solutionsWithTvlbyte, err := json.Marshal(solutionsWithTvl)
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
	}

	projectByte, err := json.Marshal(projects)
	if err != nil {

		c.JSON(http.StatusInternalServerError, nil)
	}

	c.JSON(
		http.StatusOK, gin.H{
			"chains":    string(chainsByte),
			"solutions": string(solutionsWithTvlbyte),
			"project":   string(projectByte),
		})
}
