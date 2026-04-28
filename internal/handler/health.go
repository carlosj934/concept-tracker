package handler

import(
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHealthRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})
}
