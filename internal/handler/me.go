package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterMeRoutes(router *gin.RouterGroup) {
	router.GET("/me", func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": userID,
		})
	})
}
