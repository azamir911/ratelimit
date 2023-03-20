package api

import (
	"RateLimit/api_common"
	"RateLimit/rate_limit"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
	"time"
)

func getRateLimitMiddlewareHandlerFunc(rateLimiter rate_limit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		counter, allow, err := rateLimiter.Allow(c.Request.URL)
		if err != nil {
			log.Printf("Failed to validate request url '%s'", c.Request.URL)
			c.Abort()
		}

		if allow {
			log.Printf("URL %s is reported, count=%d, not blocked", c.Request.URL, counter)
		} else {
			log.Printf("URL %s is reported, count=%d, blocked", c.Request.URL, counter)
			api_common.WriteResponse(c.Writer, http.StatusTooManyRequests, "", errors.New("URL retch the threshold"))
			c.Abort()
		}
	}
}

type reportInput struct {
	Url string `json:"url"`
}

type reportOutput struct {
	Block bool `json:"block"`
}

func GetRateLimitReportHandlerFunc(rateLimiter rate_limit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input reportInput
		if err := c.BindJSON(&input); err != nil {
			log.Println("Failed parsing input. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		theUrl, err := url.Parse(input.Url)
		if err != nil {
			log.Println("Failed parsing URL. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		allow, err := rateLimiter.State(theUrl)
		if err != nil {
			log.Println("Failed get URL state. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		output := reportOutput{Block: !allow}

		bytes, err := json.Marshal(output)
		if err != nil {
			log.Println("Failed marshaling output. Err '%w'", err)
			c.Status(http.StatusInternalServerError)
		}

		c.Data(http.StatusOK, "application/json", bytes)
	}
}

func GetRateLimitReportHandlerFunc2(rateLimiter rate_limit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input reportInput
		if err := c.BindJSON(&input); err != nil {
			log.Println("Failed parsing input. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		url, err := url.Parse(input.Url)
		if err != nil {
			log.Println("Failed parsing URL. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		//counter, allow, err := rateLimiter.Allow(c.Request.URL)
		counter, allow, err := rateLimiter.Allow(url)
		if err != nil {
			log.Println("Failed get URL state. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		_ = fmt.Sprint(time.Now().UTC().Format("15:04:05"))
		if allow {
			log.Printf("URL %s is reported, count=%d, not blocked", url, counter)
		} else {
			log.Printf("URL %s is reported, count=%d, blocked", url, counter)
			//api_common.WriteResponse(c.Writer, http.StatusTooManyRequests, "", errors.New("URL retch the threshold"))
			//c.Abort()
		}

		output := reportOutput{Block: !allow}

		bytes, err := json.Marshal(output)
		if err != nil {
			log.Println("Failed marshaling output. Err '%w'", err)
			c.Status(http.StatusInternalServerError)
		}

		c.Data(http.StatusOK, "application/json", bytes)
	}
}
