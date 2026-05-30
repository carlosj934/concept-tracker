package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/service"
)

func RegisterReminderRoutes(router *gin.RouterGroup, h *ReminderHandler) {
	router.GET("/concepts/:id/reminders", h.ListConceptReminders)
	router.POST("/concepts/:id/reminders", h.Create)
	router.PATCH("/concepts/:id/reminders/:rid", h.Update)
	router.DELETE("/concepts/:id/reminders/:rid", h.Delete)
}

type ReminderHandler struct {
	service service.ReminderService
}

func NewReminderHandler(service service.ReminderService) *ReminderHandler {
	return &ReminderHandler{
		service: service,
	}
}

func (h *ReminderHandler) ListConceptReminders(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	l, err := h.service.ListConceptReminders(c, userID, c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": l,
	})
}

type createReminderRequest struct {
	Message     string     `json:"message" binding:"required"`
	IsRecurring bool       `json:"is_recurring"`
	CronExpr    *string    `json:"cron_expr"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

func (h *ReminderHandler) Create(c *gin.Context) {
	var reminder domain.Reminder
	var reminderRequest createReminderRequest

	j := c.ShouldBindJSON(&reminderRequest)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	if reminderRequest.IsRecurring && reminderRequest.CronExpr == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "cron_expr is required for recurring reminders",
			},
		})
		return
	}

	reminder = domain.Reminder{
		Message:     reminderRequest.Message,
		IsRecurring: reminderRequest.IsRecurring,
		CronExpr:    reminderRequest.CronExpr,
		ScheduledAt: reminderRequest.ScheduledAt,
	}

	userID, ok := getUserID(c)
	if !ok {
		return
	}

	create, err := h.service.Create(c, c.Param("id"), userID, reminder)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": create,
	})
}

func (h *ReminderHandler) Update(c *gin.Context) {
	var updateReminder domain.UpdateReminderParams

	j := c.ShouldBindJSON(&updateReminder)
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

	u, err := h.service.Update(c, userID, c.Param("rid"), updateReminder)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": u,
	})
}

func (h *ReminderHandler) Delete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c, userID, c.Param("rid")); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
