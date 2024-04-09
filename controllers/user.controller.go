package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Origho-precious/url-shortener/go/models"
	"github.com/Origho-precious/url-shortener/go/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleSignup(c *gin.Context, us *models.UserService) {
	var reqBody struct {
		Email    string `json:"email" binding:"required,email"`
		FullName string `json:"fullName" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.Capitalise(err.Error())})
		return
	} else if len(reqBody.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password must be at least 8 characters long",
		})
		return
	}

	userData := models.User{
		Email:    strings.ToLower(reqBody.Email),
		Password: reqBody.Password,
		FullName: reqBody.FullName,
	}

	us.User = userData

	createdUser, err := us.CreateUser()
	if err != nil {
		log.Println(err)

		var statusCode int

		if err.Error() == "internal server error" {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	idString := createdUser.ID.Hex()

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"response": map[string]any{
			"id":            idString,
			"email":         createdUser.Email,
			"fullName":      createdUser.FullName,
			"authToken":     createdUser.AuthToken,
			"emailVerified": createdUser.EmailVerified,
		},
	})
}

func HandleLogin(c *gin.Context, us *models.UserService) {
	var reqBody struct {
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	us.User = models.User{
		Email:    strings.ToLower(reqBody.Email),
		Password: reqBody.Password,
	}

	loggedInUser, err := us.AuthenticateUser()
	if err != nil {
		log.Println(err)

		var statusCode int

		if err.Error() == "internal server error" {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	idString := loggedInUser.ID.Hex()

	c.JSON(http.StatusOK, gin.H{
		"message": "User successfully authenticated",
		"response": map[string]any{
			"id":            idString,
			"email":         loggedInUser.Email,
			"fullName":      loggedInUser.FullName,
			"authToken":     loggedInUser.AuthToken,
			"emailVerified": loggedInUser.EmailVerified,
		},
	})
}

func HandleEmailVerification(c *gin.Context, us *models.UserService) {
	type ReqBody struct {
		Token string `json:"token" binding:"required"`
	}

	var reqBody ReqBody

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token not provided"})
		return
	}

	userId := c.MustGet("userId").(string)
	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	us.User.ID = objectID

	err = us.VerifyUserEmail(reqBody.Token)
	if err != nil {
		var statusCode int

		if err.Error() == "internal server error" {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Your email has been verified, you can now shorten your URLs.",
	})
}

func HandleForgotPassword(c *gin.Context, us *models.UserService) {
	var reqBody struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	us.User.Email = reqBody.Email

	err := us.ForgotPassword()
	errMsg := fmt.Errorf(
		"no account associated with email address: %s", reqBody.Email,
	)
	if err != nil {
		var statusCode int

		if err.Error() == errMsg.Error() {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusInternalServerError
		}

		c.JSON(statusCode, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Check your email for link to reset your password",
	})
}

func HandlePasswordReset(c *gin.Context, us *models.UserService) {
	var reqBody struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.Capitalise(err.Error())})
		return
	} else if len(reqBody.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password must be at least 8 characters long",
		})
		return
	}

	us.User.Password = reqBody.Password

	err := us.ResetPassword(c.Query("token"))
	if err != nil {
		var statusCode int

		if err.Error() == "internal server error" {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Password reset successful"})
}

func HandleEmailVerificationTokenResend(
	c *gin.Context, us *models.UserService,
) {
	userEmail := c.MustGet("userEmail").(string)

	userId := c.MustGet("userId").(string)
	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	us.User.ID = objectID
	us.User.Email = userEmail

	err = us.ResendEmailVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": utils.Capitalise(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Check your email address for verification token.",
	})
}

func GetUserProfile(c *gin.Context, us *models.UserService) {
	userId := c.MustGet("userId").(string)
	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	us.User.ID = objectID

	userData, err := us.GetUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": utils.Capitalise(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successful.",
		"response": map[string]any{
			"id":            userId,
			"email":         userData.Email,
			"fullName":      userData.FullName,
			"createdAt":     userData.CreatedAt,
			"emailVerified": userData.EmailVerified,
		},
	})
}

func HandleUserFullNameEdit(c *gin.Context, us *models.UserService) {
	var reqBody struct {
		FullName string `json:"fullName" binding:"required"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	if len(reqBody.FullName) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty string is not a valid fullName value",
		})
		return
	}

	us.User.FullName = reqBody.FullName

	userId := c.MustGet("userId").(string)
	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	us.User.ID = objectID

	userData, err := us.UpdateUserFullName()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.Capitalise(err.Error())})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"message": "Fullname updated successfully.",
		"response": map[string]any{
			"id":            userId,
			"email":         userData.Email,
			"fullName":      userData.FullName,
			"createdAt":     userData.CreatedAt,
			"emailVerified": userData.EmailVerified,
		},
	})
}
