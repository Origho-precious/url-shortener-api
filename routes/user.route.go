package routes

import (
	"github.com/Origho-precious/url-shortener/go/controllers"
	"github.com/Origho-precious/url-shortener/go/middlewares"
	"github.com/Origho-precious/url-shortener/go/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserRouter(r *gin.Engine, DB *mongo.Database) {
	userCollection := DB.Collection("Users")
	forgotPasswordCollection := DB.Collection("ForgotPassword")
	verificationTokenCollection := DB.Collection("VerificationToken")

	userService := &models.UserService{
		UserCollection:              userCollection,
		ForgotPasswordCollection:    forgotPasswordCollection,
		VerificationTokenCollection: verificationTokenCollection,
	}

	validateAuthToken := middlewares.ValidateAuthToken

	usersRouter := r.Group("/v1/api/users")
	{
		// Route for user signup
		usersRouter.POST("/", func(c *gin.Context) {
			controllers.HandleSignup(c, userService)
		})

		// Route for user login
		usersRouter.POST("/login", func(c *gin.Context) {
			controllers.HandleLogin(c, userService)
		})

		// Route for email verification
		usersRouter.POST("/verify", validateAuthToken(), func(c *gin.Context) {
			controllers.HandleEmailVerification(c, userService)
		})

		// Route for resending email verification token
		usersRouter.GET("/resend-verification-token", validateAuthToken(),
			func(c *gin.Context) {
				controllers.HandleEmailVerificationTokenResend(c, userService)
			},
		)

		// Route for initiating forgot password flow
		usersRouter.POST("/forgot-password", func(c *gin.Context) {
			controllers.HandleForgotPassword(c, userService)
		})

		// Route for resetting password
		usersRouter.PATCH("/reset-password",
			middlewares.ValidateResetPasswordToken(forgotPasswordCollection),
			func(c *gin.Context) {
				controllers.HandlePasswordReset(c, userService)
			},
		)

		// Route for getting user profile
		usersRouter.GET("/me", validateAuthToken(), func(c *gin.Context) {
			controllers.GetUserProfile(c, userService)
		})

		// Route for editing full name
		usersRouter.PATCH("/edit", validateAuthToken(), func(c *gin.Context) {
			controllers.HandleUserFullNameEdit(c, userService)
		})
	}
}
