package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/controllerloops"
	"github.com/l2planet/l2planet-api/src/controllers"
	"github.com/l2planet/l2planet-api/src/models"
)

func main() {

	db.GetClient().AutoMigrate(&models.Token{}, &models.Solution{}, &models.Bridge{} /*&models.Balance{},*/, &models.Newsletter{}, &models.Price{}, &models.Tvl{}, &models.Chain{}, &models.Project{})
	db.GetClient().SyncDb()
	ticker := time.NewTicker(15 * time.Minute)
	done := make(chan bool)
	err := controllerloops.CalculateTvl()
	fmt.Println("error while calculating tvls", err)
	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				err := controllerloops.CalculateTvl()
				fmt.Println("error while calculating tvls", err)
			}
		}
	}()

	r := gin.Default()

	r.GET("/info", controllers.Info)
	r.POST("/solution", controllers.NewSolution)
	r.POST("/project", controllers.NewProject)
	r.POST("/chain", controllers.NewChain)
	r.POST("/newsletter", controllers.NewNewsletter)

	r.Run(":8080")

}
