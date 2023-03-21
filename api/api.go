package api

import (
	"RateLimit/enviroment"
	rateLimit "RateLimit/rate_limit"
	"github.com/gin-gonic/gin"
)

func Serve() error {
	threshold, err := enviroment.GetEnvAsInt("threshold")
	if err != nil {
		return err
	}

	ttl, err := enviroment.GetEnvAsInt("ttl")
	if err != nil {
		return err
	}

	rateLimiter := rateLimit.New(threshold, ttl)
	report := GetRateLimitReportHandlerFunc(rateLimiter)

	engine := gin.Default()
	engine.POST("/report", report)

	if err = engine.Run("localhost:8080"); err != nil {
		return err
	}

	return nil
}
