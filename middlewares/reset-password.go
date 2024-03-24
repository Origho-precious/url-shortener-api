package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ValidateResetPasswordToken(rpwdCol *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token query param is required",
			})
			c.Abort()
			return
		}

		// Convert the token to ObjectID
		objectID, err := primitive.ObjectIDFromHex(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			c.Abort()
			return
		}

		// Check for record associated to this token in DB
		res := rpwdCol.FindOne(ctx, bson.M{"_id": objectID})
		if res.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid or expired reset password token",
			})
			c.Abort()
			return
		} else if res.Err() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
