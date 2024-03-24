package main

import (
	"net/http"

	"github.com/Origho-precious/url-shortener/go/configs"
	"github.com/Origho-precious/url-shortener/go/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	database, err := configs.ConnectDB()
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello there :)",
		})
	})

	routes.UserRouter(r, database)

	routes.UrlRouter(r, database)

	http.ListenAndServe(":5500", r)
}
