package controllers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Origho-precious/url-shortener/go/configs"
	"github.com/Origho-precious/url-shortener/go/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validateURL(url string) bool {
	domainRegex := regexp.MustCompile(
		`^(https?:\/\/)?([\w-]+\.)*([\w-]+\.[\w-]{2,})(\/[\w-./?%&=]*)?$`,
	)
	return domainRegex.MatchString(url)
}

func HandleCreateShortUrl(
	c *gin.Context,
	urlS *models.UrlService,
	us *models.UserService,
) {
	var reqBody []struct {
		Url        string `json:"url" binding:"required"`
		Alias      string `json:"alias"`
		ExpiryDate string `json:"expiryDate"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		if strings.Contains(
			err.Error(),
			"json: cannot unmarshal object into Go value of type []struct",
		) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "request body must be an array of objects with properties: url (required), alias (optional), expiryDate(optional)",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("userId").(string)
	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	urlS.Url.UserId = objectID

	us.User.ID = objectID

	userData, err := us.GetUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !userData.EmailVerified {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user's email is not yet verified",
		})
		return
	}

	var responses []map[string]string
	for _, item := range reqBody {
		if !validateURL(item.Url) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "url is invalid: " + item.Url,
			})
			return
		}

		if item.Alias != "" && len(item.Alias) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "alias needs to be at least 4 characters long: " + item.Alias,
			})
			return
		}

		urlS.Url.UserId = objectID
		urlS.Url.OriginalUrl = strings.ToLower(item.Url)

		if item.ExpiryDate != "" {
			layout := "02-01-2006"
			parsedDate, err := time.Parse(layout, item.ExpiryDate)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				return
			}

			urlS.Url.ExpiresAt = parsedDate
		} else {
			urlS.Url.ExpiresAt = time.Time{}
		}

		res, err := urlS.CreateShortUrl(item.Alias)
		if err != nil {
			var statusCode int

			if err.Error() == "internal server error" {
				statusCode = http.StatusInternalServerError
			} else {
				statusCode = http.StatusConflict
			}

			c.JSON(statusCode, gin.H{"error": err.Error()})

			return
		}

		responses = append(responses, res)
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    responses,
		"message": "URLs shortened successfully",
	})
}

func HandleUrlDelete(c *gin.Context,
	urlS *models.UrlService,
	us *models.UserService,
) {
	userId := c.MustGet("userId").(string)
	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	urlID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	urlS.Url.ID = urlID
	us.User.ID = objectID
	urlS.Url.UserId = objectID

	userData, err := us.GetUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !userData.EmailVerified {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user's email is not yet verified",
		})
		return
	}

	err = urlS.DeleteUrl()
	if err != nil {
		var statusCode int

		if err.Error() == "internal server error" {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Url deleted successfully",
	})
}

func GetUrlsByUserID(c *gin.Context, urlS *models.UrlService) {
	const (
		defaultPage     = 1
		defaultPageSize = 10
	)

	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultPageSize)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	userId := c.MustGet("userId").(string)
	objectID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	urlS.Url.UserId = objectID

	data, total, err := urlS.GetUrlsByUser(page, limit)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cfg, err := configs.LoadEnvs()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	var response []any
	for _, urlRecord := range data {
		shortUrl := fmt.Sprintf(
			"%s/%s", cfg.URL_REDIRECT_PREFIX, urlRecord.ShortUrlSlug,
		)

		urlInfo := make(map[string]interface{})
		urlInfo["id"] = urlRecord.ID.Hex()
		urlInfo["shortUrl"] = shortUrl
		urlInfo["createdAt"] = urlRecord.CreatedAt
		urlInfo["expiresAt"] = urlRecord.ExpiresAt
		urlInfo["visitCount"] = urlRecord.VisitCount
		urlInfo["originalUrl"] = urlRecord.OriginalUrl
		urlInfo["customAlias"] = urlRecord.CustomAlias
		urlInfo["qrCodeImageUrl"] = urlRecord.QRCodeImageUrl

		if urlRecord.LastVisitedAt.IsZero() {
			urlInfo["lastVisitedAt"] = nil
		} else {
			urlInfo["lastVisitedAt"] = urlRecord.LastVisitedAt
		}

		if urlRecord.ExpiresAt.IsZero() {
			urlInfo["expiresAt"] = nil
		} else {
			urlInfo["expiresAt"] = urlRecord.ExpiresAt
		}

		response = append(response, urlInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"total":   total,
		"message": "Urls generated by user",
	})
}
