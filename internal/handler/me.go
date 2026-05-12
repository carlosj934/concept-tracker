package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterMeRoutes(router *gin.RouterGroup) {
	router.GET("/me", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "forbidden",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": userID,
		})
	})
}
