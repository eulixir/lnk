package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type URLsHandler struct {
	logger *zap.Logger
}

func NewURLsHandler(logger *zap.Logger) *URLsHandler {
	return &URLsHandler{
		logger: logger,
	}
}

func (h *URLsHandler) CreateURL(c *gin.Context) {
	h.logger.Info("CreateURL called")
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (h *URLsHandler) GetURL(c *gin.Context) {
	shortURL := c.Param("short_url")
	h.logger.Info("GetURL called", zap.String("short_url", shortURL))
	c.JSON(http.StatusOK, gin.H{"message": "OK", "short_url": shortURL})
}
