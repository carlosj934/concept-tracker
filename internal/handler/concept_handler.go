package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/service"
)

func RegisterConceptRoutes(router *gin.RouterGroup, h *ConceptHandler) {
	router.GET("/concepts", h.ListRoots)
	router.GET("/concepts/:id", h.GetByID)
	router.GET("/concepts/:id/tree", h.GetSubtree)
	router.POST("/concepts", h.Create)
	router.PATCH("/concepts/:id", h.Update)
	router.PATCH("/concepts/:id/move", h.Move)
	router.DELETE("/concepts/:id", h.Delete)
}

type ConceptHandler struct {
	service service.ConceptService
}

func NewConceptHandler(service service.ConceptService) *ConceptHandler {
	return &ConceptHandler{
		service: service,
	}
}

func getUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "forbidden",
			},
		})
		return "", false
	}

	return userID.(string), true
}

func (h *ConceptHandler) ListRoots(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	l, err := h.service.ListRoots(c, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": l,
	})
}

func (h *ConceptHandler) GetByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	g, err := h.service.GetByID(c, userID, c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": g,
	})
}

func (h *ConceptHandler) GetSubtree(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	g, err := h.service.GetSubtree(c, userID, c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": g,
	})
}

type createConceptRequest struct {
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func (h *ConceptHandler) Create(c *gin.Context) {
	var concept domain.Concept
	var conceptRequest createConceptRequest

	j := c.ShouldBindJSON(&conceptRequest)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	concept = domain.Concept{
		ParentID:    conceptRequest.ParentID,
		Name:        conceptRequest.Name,
		Description: conceptRequest.Description,
	}

	userID, ok := getUserID(c)
	if !ok {
		return
	}

	create, err := h.service.Create(c, userID, concept)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": create,
	})
}

type updateConceptRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *ConceptHandler) Update(c *gin.Context) {
	var updateConcept updateConceptRequest

	j := c.ShouldBindJSON(&updateConcept)
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

	u, err := h.service.Update(c, userID, c.Param("id"), updateConcept.Name, updateConcept.Description)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": u,
	})
}

type moveConceptRequest struct {
	NewParentID *string `json:"newParentID"`
}

func (h *ConceptHandler) Move(c *gin.Context) {
	var moveConcept moveConceptRequest

	j := c.ShouldBindJSON(&moveConcept)
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

	m := h.service.Move(c, userID, c.Param("id"), moveConcept.NewParentID)
	if m != nil {
		handleError(c, m)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ConceptHandler) Delete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	d := h.service.Delete(c, userID, c.Param("id"))
	if d != nil {
		handleError(c, d)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "not found",
			},
		})
	} else {
		log.Printf("error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "internal server error",
			},
		})
	}
}
