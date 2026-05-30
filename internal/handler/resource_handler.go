package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/service"
)

func RegisterResourceRoutes(router *gin.RouterGroup, h *ResourceHandler) {
	router.GET("/concepts/:id/resources", h.ListConceptResources)
	router.POST("/concepts/:id/resources", h.Create)
	router.PATCH("/concepts/:id/resources/:rid", h.Update)
	router.DELETE("/concepts/:id/resources/:rid", h.Delete)
}

type ResourceHandler struct {
	service service.ResourceService
}

func NewResourceHandler(service service.ResourceService) *ResourceHandler {
	return &ResourceHandler{
		service: service,
	}
}

func (h *ResourceHandler) ListConceptResources(c *gin.Context) {
	l, err := h.service.ListConceptResources(c, getUserID(c), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": l,
	})
}

type createResourceRequest struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"external_id"`
	URL        string `json:"url"`
	Title      string `json:"title"`
}

func (h *ResourceHandler) Create(c *gin.Context) {
	var resource domain.ConceptResource
	var resourceRequest createResourceRequest

	j := c.ShouldBindJSON(&resourceRequest)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	resource = domain.ConceptResource{
		Provider:   resourceRequest.Provider,
		ExternalID: resourceRequest.ExternalID,
		URL:        resourceRequest.URL,
		Title:      resourceRequest.Title,
	}

	create, err := h.service.Create(c, getUserID(c), c.Param("id"), resource)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": create,
	})
}

type updateResourceRequest struct {
	URL   *string `json:"url"`
	Title *string `json:"title"`
}

func (h *ResourceHandler) Update(c *gin.Context) {
	var updateResource updateResourceRequest

	j := c.ShouldBindJSON(&updateResource)
	if j != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "bad request",
			},
		})
		return
	}

	u, err := h.service.Update(c, getUserID(c), c.Param("rid"), updateResource.URL, updateResource.Title)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": u,
	})
}

func (h *ResourceHandler) Delete(c *gin.Context) {
	d := h.service.Delete(c, getUserID(c), c.Param("rid"))
	if d != nil {
		handleError(c, d)
		return
	}

	c.Status(http.StatusNoContent)
}
