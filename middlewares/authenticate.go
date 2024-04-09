package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Origho-precious/url-shortener/go/configs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func isTimestampBeforeToday(timestamp float64) bool {
	timestampInt := int64(timestamp)

	date := time.Unix(timestampInt, 0).UTC().Truncate(24 * time.Hour)

	today := time.Now().UTC().Truncate(24 * time.Hour)

	return today.Before(date)
}

func ValidateAuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := configs.LoadEnvs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			c.Abort()
			return
		}

		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No auth token provided"})
			c.Abort()
			return
		}

		authToken := strings.Split(tokenHeader, " ")[1]

		token, err := jwt.Parse(authToken, func(token *jwt.Token) (any, error) {
			return []byte(cfg.JWT_SECRET), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired auth token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			c.Set("userId", claims["userId"])
			c.Set("userEmail", claims["userEmail"])

			exp := claims["exp"].(float64)
			if exp == 0 {
				fmt.Println("exp claim is invalid", exp)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
				c.Abort()
				return
			} else {
				notExpired := isTimestampBeforeToday(exp)
				if notExpired {
					c.Next()
				} else {
					fmt.Println("Expired auth token => ", exp)
					c.JSON(http.StatusBadRequest, gin.H{"error": "auth token expired"})
					c.Abort()
					return
				}
			}
		} else {
			fmt.Println("Invalid auth token claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ValidateAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg, err := configs.LoadEnvs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			c.Abort()
			return
		}

		tokenHeader := c.GetHeader("x-api-key")
		if tokenHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No api-key provided"})
			c.Abort()
			return
		}

		if tokenHeader != cfg.API_KEY {
			c.JSON(http.StatusForbidden, gin.H{"error": ""})
			c.Abort()
			return
		}

		c.Next()
	}
}
