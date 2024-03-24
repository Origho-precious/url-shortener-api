package routes

import (
	"github.com/Origho-precious/url-shortener/go/controllers"
	"github.com/Origho-precious/url-shortener/go/middlewares"
	"github.com/Origho-precious/url-shortener/go/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func UrlRouter(r *gin.Engine, DB *mongo.Database) {
	urlCollection := DB.Collection("Urls")
	userCollection := DB.Collection("Users")
	visitCollection := DB.Collection("Visits")

	urlService := &models.UrlService{
		UrlCollection:   urlCollection,
		VisitCollection: visitCollection,
	}

	userService := &models.UserService{
		UserCollection: userCollection,
	}

	validateAuthToken := middlewares.ValidateAuthToken

	r.GET("/redirect/:slug", func(c *gin.Context) {
		controllers.RedirectToLongUrl(c, urlService)
	})

	router := r.Group("/v1/api/urls")
	{
		router.POST("/", validateAuthToken(), func(c *gin.Context) {
			controllers.HandleCreateShortUrl(c, urlService, userService)
		})

		router.GET("/", validateAuthToken(), func(c *gin.Context) {
			controllers.GetUrlsByUserID(c, urlService)
		})

		router.DELETE("/:id/delete", validateAuthToken(), func(c *gin.Context) {
			controllers.HandleUrlDelete(c, urlService, userService)
		})
	}

}
