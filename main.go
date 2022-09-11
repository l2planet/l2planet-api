package main

import (
	"github.com/gin-gonic/gin"
	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/controllers"
	"github.com/l2planet/l2planet-api/src/models"
)

func main() {

	db.GetClient().AutoMigrate(&models.Token{}, &models.Solution{}, &models.Bridge{} /*&models.Balance{},*/, &models.Newsletter{}, &models.Price{}, &models.Tvl{}, &models.Chain{}, &models.Project{})
	//db.GetClient().GetAllSolutionsWithTvl()

	//db.GetClient().SyncDb()
	/*start := time.Now()
	controllerloops.CalculateTvl()
	elapsed := time.Since(start)
	fmt.Printf("calculate tvl took %s", elapsed)*/

	r := gin.Default()

	r.GET("/info", controllers.Info)
	r.POST("/solution", controllers.NewSolution)
	r.POST("/project", controllers.NewProject)
	r.POST("/chain", controllers.NewChain)
	r.POST("/newsletter", controllers.NewNewsletter)

	r.Run()

}
