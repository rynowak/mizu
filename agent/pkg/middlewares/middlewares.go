package middlewares

import (
	"mizuserver/pkg/config"
	"mizuserver/pkg/providers"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/up9inc/mizu/shared/logger"
)

const cachedValidTokensRetainmentTime = time.Minute * 1

var cachedValidTokens = cache.New(cachedValidTokensRetainmentTime, cachedValidTokensRetainmentTime)

func RequiresAuth(c *gin.Context) {
	// auth is irrelevant for ephermal mizu instances
	if config.Config.IsEphermeral {
		c.Next()
		return
	}

	bearerToken := c.GetHeader("Authorization")

	if bearerToken == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header"})
		return
	}
	if !strings.HasPrefix(bearerToken, "Bearer ") {
		c.AbortWithStatusJSON(401, gin.H{"error": "authorization header must be a bearer token"})
		return
	}

	token := strings.Split(bearerToken, " ")[1]
	if token == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "bearer token is empty"})
		return
	}

	if _, isTokenCached := cachedValidTokens.Get(token); isTokenCached {
		c.Next()
		return
	}

	if isTokenValid, err := providers.VerifyToken(token, c.Request.Context()); err != nil {
		logger.Log.Errorf("error verifying token %s", err)
		c.AbortWithStatusJSON(401, gin.H{"error": "unknown auth error occured"})
		return
	} else if !isTokenValid {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
		return
	}

	cachedValidTokens.Set(token, true, cachedValidTokensRetainmentTime)

	c.Next()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}