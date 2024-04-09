package main

import (
	"net/http"
	"time"

	"github.com/Origho-precious/url-shortener/go/configs"
	"github.com/Origho-precious/url-shortener/go/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var methods = []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}

func main() {
	database, err := configs.ConnectDB()
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		MaxAge:           12 * time.Hour,
		AllowMethods:     methods,
		AllowOrigins:     []string{"http://localhost:3001"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "x-api-key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello there :)"})
	})

	routes.UserRouter(r, database)

	routes.UrlRouter(r, database)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Oops, not Found :("})
	})

	http.ListenAndServe(":5500", r)
}
