package main

import (
	"fmt"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/l2planet/l2planet-api/src/clients/db"
	"github.com/l2planet/l2planet-api/src/consts"
	"github.com/l2planet/l2planet-api/src/controllerloops"
	"github.com/l2planet/l2planet-api/src/controllers"
	"github.com/l2planet/l2planet-api/src/models"
)

const (
	identityKey  = "id"
	identityName = "name"
)

func main() {

	db.GetClient().AutoMigrate(&models.Token{}, &models.Solution{}, &models.Bridge{} /*&models.Balance{},*/, &models.Users{}, &models.Newsletter{}, &models.Price{}, &models.Tvl{}, &models.Chain{}, &models.Project{})
	//db.GetClient().SyncDb()
	ticker := time.NewTicker(15 * time.Minute)
	done := make(chan bool)
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

	// the jwt middleware
	authMiddleware := &jwt.GinJWTMiddleware{
		Realm:       "Dev",
		Key:         []byte(consts.JwtSecret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			v, ok := data.(*models.Users)
			if ok {
				return jwt.MapClaims{
					identityKey:  v.ID,
					identityName: v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.Users{
				Username: claims[identityName].(string),
				Model: gorm.Model{
					ID: uint(claims[identityKey].(float64)),
				},
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals models.Users
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			userDb, err := db.GetClient().GetUser(loginVals.Username)
			if err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			if userDb == nil {
				return nil, jwt.ErrFailedAuthentication
			}

			if err := bcrypt.CompareHashAndPassword([]byte(userDb.Password), []byte(loginVals.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			return userDb, nil

		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if _, ok := data.(*models.Users); ok {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	}
	err := authMiddleware.MiddlewareInit()

	if err != nil {
		fmt.Println("authMiddleware.MiddlewareInit() Error:" + err.Error())
		panic(err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "access-control-allow-origin", "access-control-allow-headers", "Authorization"},
	}))

	r.POST("/login", authMiddleware.LoginHandler)
	r.POST("/register", controllers.Register)
	r.GET("/info", controllers.Info)

	auth := r.Group("/auth")
	auth.Use(authMiddleware.MiddlewareFunc())
	auth.POST("/solution", controllers.NewSolution)
	auth.POST("/project", controllers.NewProject)
	auth.POST("/chain", controllers.NewChain)
	auth.POST("/newsletter", controllers.NewNewsletter)

	r.Run(":8080")

}
