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
	l, err := h.service.ListConceptReminders(c, getUserID(c), c.Param("id"))
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

	reminder = domain.Reminder{
		Message:     reminderRequest.Message,
		IsRecurring: reminderRequest.IsRecurring,
		CronExpr:    reminderRequest.CronExpr,
		ScheduledAt: reminderRequest.ScheduledAt,
	}

	create, err := h.service.Create(c, c.Param("id"), getUserID(c), reminder)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": create,
	})
}

func (h *ReminderHandler) Update(c *gin.Context) {
	var updateReminder service.UpdateReminderParams

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

	u, err := h.service.Update(c, getUserID(c), c.Param("rid"), updateReminder)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": u,
	})
}

func (h *ReminderHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c, getUserID(c), c.Param("rid")); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
