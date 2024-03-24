package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Origho-precious/url-shortener/go/models"
	"github.com/gin-gonic/gin"
)

func getDeviceTypeFromUserAgent(userAgent string) string {
	if strings.Contains(strings.ToLower(userAgent), "macintosh") {
		return "MacOS"
	} else if strings.Contains(strings.ToLower(userAgent), "windows") {
		return "Windows"
	} else if strings.Contains(strings.ToLower(userAgent), "linux") {
		return "Linux"
	} else if strings.Contains(strings.ToLower(userAgent), "iphone") ||
		strings.Contains(strings.ToLower(userAgent), "ipad") {
		return "iOS"
	} else if strings.Contains(strings.ToLower(userAgent), "android") {
		return "Android"
	} else {
		return "Unknown"
	}
}

func getBrowserFromUserAgent(userAgent string) string {
	if strings.Contains(strings.ToLower(userAgent), "chrome") {
		return "Chrome"
	} else if strings.Contains(strings.ToLower(userAgent), "firefox") {
		return "Firefox"
	} else if strings.Contains(strings.ToLower(userAgent), "safari") {
		return "Safari"
	} else if strings.Contains(strings.ToLower(userAgent), "edge") {
		return "Edge"
	} else {
		return "Other"
	}
}

func RedirectToLongUrl(c *gin.Context, urlS *models.UrlService) {
	urlSlug := c.Param("slug")
	urlS.Url.ShortUrlSlug = urlSlug

	response, err := urlS.GetOriginalUrl()
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

	referrer := c.Request.Referer()
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()
	browser := getBrowserFromUserAgent(userAgent)
	deviceType := getDeviceTypeFromUserAgent(userAgent)

	urlS.Visit.UrlId = response.ID
	urlS.Visit.Browser = browser
	urlS.Visit.Referrer = referrer
	urlS.Visit.IPAddress = ipAddress
	urlS.Visit.DeviceType = deviceType
	urlS.Visit.VisitedAt = time.Now()

	go func() {
		err := urlS.SaveClickAnalytics()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Visit analytic saved.")
	}()

	fmt.Println("Redirecting to: ", response.OriginalUrl)

	c.Redirect(http.StatusTemporaryRedirect, response.OriginalUrl)
}
