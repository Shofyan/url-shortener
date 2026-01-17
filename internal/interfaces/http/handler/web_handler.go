package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebHandler handles web UI requests.
type WebHandler struct{}

// NewWebHandler creates a new WebHandler.
func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

// Index serves the main page.
func (h *WebHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "URL Shortener",
	})
}
