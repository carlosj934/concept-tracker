package handler

import (
	"net/http"

	"concept-tracker/internal/service"
	"concept-tracker/internal/domain"

	"github.com/gin-gonic/gin"
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

func NewConceptHandler (service service.ConceptService) *ConceptHandler {
	return &ConceptHandler{
		service: service,
	}
}

func getUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code": "FORBIDDEN",
				"message": "forbidden",
			},
		})
		return ""
	}

	return userID.(string)
}

func (h *ConceptHandler) ListRoots(c *gin.Context) {

	l, err := h.service.ListRoots(c, getUserID(c))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": l,
	})
}

func (h *ConceptHandler) GetByID(c *gin.Context) {
	g, err := h.service.GetByID(c, getUserID(c), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": g,
	})

}

func (h *ConceptHandler) GetSubtree(c *gin.Context) {
	g, err := h.service.GetSubtree(c, getUserID(c), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": g,
	})
}

func (h *ConceptHandler) Create(c *gin.Context) {
	var concept domain.Concept

	j := c.ShouldBindJSON(&concept)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code": "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	create, err := h.service.Create(c, getUserID(c), concept)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": create,
	})
}

type updateConceptRequest struct {
	Name string `json:"name"`
	Description *string `json:"description"`
}

func (h *ConceptHandler) Update(c *gin.Context) {
	var updateConcept updateConceptRequest

	j := c.ShouldBindJSON(&updateConcept)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code": "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}
	
	u, err := h.service.Update(c, getUserID(c), c.Param("id"), updateConcept.Name, updateConcept.Description)
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
				"code": "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	m := h.service.Move(c, getUserID(c), c.Param("id"), moveConcept.NewParentID)
	if m != nil {
		handleError(c, m)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ConceptHandler) Delete(c *gin.Context) {
	d := h.service.Delete(c, getUserID(c), c.Param("id"))
	if d != nil {
		handleError(c, d)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleError(c *gin.Context, err error) {
	if err == domain.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code": "NOT_FOUND",
				"message": "not found",
			},
		})	
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": "INTERNAL_SERVER_ERROR",
				"message": "internal server error",
			},
		})
	}
}

