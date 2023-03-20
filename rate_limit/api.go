package rate_limit

//func Init()

//var rateLimiter RateLimiter

func New(threshold int, ttl int) RateLimiter {
	return newRateLimiter(threshold, ttl)
}

//func Init(threshold int, ttl int) gin.HandlerFunc {
//	rateLimiter = NewRateLimiter(threshold, ttl)
//
//	return func(c *gin.Context) {
//		//requestContext := c.Request.Context()
//		counter, allow, err := rateLimiter.Allow(c.Request.URL)
//		if err != nil {
//			log.Printf("Failed to validate request url '%s'", c.Request.URL)
//			c.Abort()
//		}
//
//		if allow {
//			log.Printf("URL %s is reported, count=%d, not blocked", c.Request.URL, counter)
//		} else {
//			log.Printf("URL %s is reported, count=%d, blocked", c.Request.URL, counter)
//			api_common.WriteResponse(c.Writer, http.StatusTooManyRequests, "", errors.New("URL retch the threshold"))
//			c.Abort()
//		}
//		//tenantId := tenants.TenantIdFromToken(requestContext)
//		//org := tenants.OrganizationFromToken(requestContext)
//		//
//		//currCorrId := uuid.New()
//		//
//		//logger := FromContext(requestContext).With().Stringer("tenantId", tenantId).Stringer(corrId, currCorrId).Logger()
//		//ctx := WithCorrId(requestContext, currCorrId)
//		//ctx = tenants.WithTenantIdAndOrg(ctx, tenantId, org)
//		//ctx = ContextWithLogger(ctx, &logger)
//
//		//newRequest := c.Request.WithContext(ctx)
//		//c.Request = newRequest
//	}
//}

//
//func Init() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		const funcName = "rolesAccess.Init"
//		requestContext := c.Request.Context()
//		writer := c.Writer
//		if err := checkRoleAccess(c, requestContext); err != nil {
//			l := logging.FromContextWithFunc(requestContext, funcName)
//			l.Err(err).Msg("Operation not allowed for current role")
//			c.Abort()
//			writer.WriteHeader(http.StatusBadRequest) // TODO: Change this to Forbidden(403 status code) after CYP-3604 is fixed
//			writer.WriteString("Operation not allowed for current role")
//			return
//		}
//	}
//}
