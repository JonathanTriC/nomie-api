package misc_handlers

import (
	"net/http"

	misc_services "github.com/JonathanTriC/nomie-api/internal/modules/misc/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   misc_services.Service
}

func NewHandler(service misc_services.Service) *Handler {
	return &Handler{
		service:   service,
	}
}

func (h *Handler) GetCategoryList(c *gin.Context) {
	categories, err := h.service.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return JSON
	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

func (h *Handler) GetAreaList(c *gin.Context) {
	areas, err := h.service.GetAreas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return JSON
	c.JSON(http.StatusOK, gin.H{
		"areas": areas,
	})
}