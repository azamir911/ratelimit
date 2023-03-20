package api

import (
	"RateLimit/enviroment"
	"RateLimit/rate_limit"
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

	//middleware := rate_limit.Init(threshold, ttl)
	rateLimiter := rate_limit.New(threshold, ttl)
	//middleware := getRateLimitMiddlewareHandlerFunc(rateLimiter)
	//report := GetRateLimitReportHandlerFunc(rateLimiter)
	report := GetRateLimitReportHandlerFunc2(rateLimiter)

	engine := gin.Default()
	//authOnly := engine.Group("/")
	//authOnly.Use(middleware)
	//engine.POST("/report", report)
	engine.POST("/report", report)

	engine.Run("localhost:8080")

	return nil
}

//func validateRateLimit(context *gin.Context) {
//	api_common.WriteResponse(context.Writer, http.StatusOK, "Welcome to the Rate Limit App!", nil)
//}
//
//func index(context *gin.Context) {
//	api_common.WriteResponse(context.Writer, http.StatusOK, "Welcome to the Rate Limit App!", nil)
//}
//
//func index2(context *gin.Context) {
//	log.Printf("Request url is '%s'\n", context.Request.URL)
//	log.Printf("Request url path is '%s'\n", context.Request.URL.Path)
//	api_common.WriteResponse(context.Writer, http.StatusOK, "Welcome to the Rate Limit App2!", nil)
//}
//
//func getReportFunc() gin.HandlerFunc {
//	return func(context *gin.Context) {
//		//ctx := context.Request.Context()
//
//		var input reportInput
//		if err := context.BindJSON(&input); err != nil {
//			log.Println("Failed parsing input. Err '%w'", err)
//			context.Status(http.StatusBadRequest)
//			return
//		}
//
//		bytes, err := json.Marshal(input)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		context.Data(http.StatusOK, "application/json", bytes)
//	}
//}
