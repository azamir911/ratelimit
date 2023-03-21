package api

import (
	"RateLimit/rate_limit"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
)

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
		} else if input.Url == "" {
			log.Println("Got empty input")
			c.Status(http.StatusBadRequest)
			return
		}

		theUrl, err := url.Parse(input.Url)
		if err != nil {
			log.Println("Failed parsing URL. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		counter, allow, err := rateLimiter.Allow(theUrl)
		if err != nil {
			log.Println("Failed get URL state. Err '%w'", err)
			c.Status(http.StatusBadRequest)
			return
		}

		//const timeFormat = "15:04:05"
		//_ = fmt.Sprint(time.Now().UTC().Format(timeFormat))
		if allow {
			log.Printf("URL %s is reported, count=%d, not blocked", theUrl, counter)
		} else {
			log.Printf("URL %s is reported, count=%d, blocked", theUrl, counter)
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
