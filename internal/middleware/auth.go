package middleware

import (
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gin-gonic/gin"
)

func ClerkAuth() gin.HandlerFunc {
	middleware := clerkhttp.WithHeaderAuthorization()
	return func(c *gin.Context) {
		middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			
			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{
						"code": "UNAUTHORIZED",
						"message": "unauthorized",
					},
				})
				return
			}

			c.Set("userID", claims.Subject)

			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}
