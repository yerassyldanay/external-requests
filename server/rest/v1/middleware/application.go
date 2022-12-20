package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ValidateHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ContentType() != "application/json" && !strings.HasPrefix(c.Request.URL.Path, "/debug") {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{
				"error": "this microservice only supports application/json",
			})
			return
		}

		c.Next()
		c.Header("Content-Type", "application/json")
		c.Header("Content-Length", fmt.Sprintf("%v", c.Request.ContentLength))
	}
}
