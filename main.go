package main

import (
	"fmt"
	"time"

	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/controllerloops"
	"github.com/l2planet/l2planet-api/src/models"
)

func main() {

	db.GetDbClient().AutoMigrate(&models.Token{}, &models.Solution{}, &models.Bridge{}, &models.Balance{}, &models.Blog{}, &models.Price{}, &models.Tvl{})
	db.GetClient().SyncDb()
	start := time.Now()
	controllerloops.CalculateTvl()
	elapsed := time.Since(start)

	fmt.Printf("calculate tvl took %s", elapsed)
}
