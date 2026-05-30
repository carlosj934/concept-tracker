package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"concept-tracker/internal/service"
)

func RegisterUserPreferencesRoutes(router *gin.RouterGroup, h *UserPreferencesHandler) {
	router.GET("/me/preferences", h.GetUserPreferences)
	router.PATCH("/me/preferences", h.Update)
}

type UserPreferencesHandler struct {
	service service.UserPreferencesService
}

func NewUserPreferencesHandler(service service.UserPreferencesService) *UserPreferencesHandler {
	return &UserPreferencesHandler{
		service: service,
	}
}

func (h *UserPreferencesHandler) GetUserPreferences(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	g, err := h.service.GetUserPreferences(c, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": g,
	})
}

type updateUserPreferencesRequest struct {
	Timezone string `json:"timezone" binding:"required"`
}

func (h *UserPreferencesHandler) Update(c *gin.Context) {
	var updateUserPreferences updateUserPreferencesRequest

	j := c.ShouldBindJSON(&updateUserPreferences)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		return
	}

	u, err := h.service.Update(c, userID, updateUserPreferences.Timezone)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": u,
	})
}
