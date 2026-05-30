package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/service"
)

func RegisterActivityLogRoutes(router *gin.RouterGroup, h *ActivityLogHandler) {
	router.GET("/concepts/:id/logs", h.List)
	router.POST("/concepts/:id/logs", h.Create)
	router.PATCH("/concepts/:id/logs/:lid", h.Update)
	router.DELETE("/concepts/:id/logs/:lid", h.Delete)
}

type ActivityLogHandler struct {
	service service.ActivityLogService
}

func NewActivityLogHandler(service service.ActivityLogService) *ActivityLogHandler {
	return &ActivityLogHandler{
		service: service,
	}
}

func (h *ActivityLogHandler) List(c *gin.Context) {
	cursor := c.Query("cursor")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	userID, ok := getUserID(c)
	if !ok {
		return
	}

	l, err := h.service.List(c, userID, c.Param("id"), cursor, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, l)
}

type createActivityLogRequest struct {
	ActivityType string    `json:"activity_type" binding:"required"`
	DurationMins *int64    `json:"duration_minutes"`
	Notes        *string   `json:"notes"`
	LoggedAt     time.Time `json:"logged_at" binding:"required"`
}

func (h *ActivityLogHandler) Create(c *gin.Context) {
	var activityLog domain.ActivityLog
	var createActivityLog createActivityLogRequest

	j := c.ShouldBindJSON(&createActivityLog)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	activityLog = domain.ActivityLog{
		ActivityType: createActivityLog.ActivityType,
		DurationMins: createActivityLog.DurationMins,
		Notes:        createActivityLog.Notes,
		LoggedAt:     createActivityLog.LoggedAt,
	}

	userID, ok := getUserID(c)
	if !ok {
		return
	}

	create, err := h.service.Create(c, userID, c.Param("id"), activityLog)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": create,
	})
}

type updateActivityLogRequest struct {
	ActivityType *string    `json:"activity_type"`
	DurationMins *int64     `json:"duration_minutes"`
	Notes        *string    `json:"notes"`
	LoggedAt     *time.Time `json:"logged_at"`
}

func (h *ActivityLogHandler) Update(c *gin.Context) {
	var updateActivityLog updateActivityLogRequest

	j := c.ShouldBindJSON(&updateActivityLog)
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

	u, err := h.service.Update(c, userID, c.Param("lid"), updateActivityLog.ActivityType, updateActivityLog.DurationMins, updateActivityLog.Notes, updateActivityLog.LoggedAt)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": u,
	})
}

func (h *ActivityLogHandler) Delete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	d := h.service.Delete(c, userID, c.Param("lid"))
	if d != nil {
		handleError(c, d)
		return
	}

	c.Status(http.StatusNoContent)
}
