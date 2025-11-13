package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateURLRequest represents the request body for creating a short URL
type CreateURLRequest struct {
	URL string `json:"url" example:"https://example.com" binding:"required"`
}

// CreateURLResponse represents the response for creating a short URL
type CreateURLResponse struct {
	ShortURL    string `json:"short_url" example:"abc123"`
	OriginalURL string `json:"original_url" example:"https://example.com"`
}

// GetURLResponse represents the response for getting a URL
type GetURLResponse struct {
	ShortURL    string `json:"short_url" example:"abc123"`
	OriginalURL string `json:"original_url" example:"https://example.com"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type URLsHandler struct {
	logger *zap.Logger
}

func NewURLsHandler(logger *zap.Logger) *URLsHandler {
	return &URLsHandler{
		logger: logger,
	}
}

// CreateURL godoc
// @Summary      Create a short URL
// @Description  Create a short URL from a long URL
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        request  body      CreateURLRequest  true  "URL to shorten"
// @Success      200      {object}  CreateURLResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /shorten [post]
func (h *URLsHandler) CreateURL(c *gin.Context) {
	var req CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info("CreateURL called", zap.String("url", req.URL))

	// TODO: Implement actual URL shortening logic
	c.JSON(http.StatusOK, CreateURLResponse{
		ShortURL:    "abc123",
		OriginalURL: req.URL,
	})
}

// GetURL godoc
// @Summary      Get original URL by short URL
// @Description  Get the original URL from a short URL
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        short_url  path      string  true  "Short URL identifier"
// @Success      301        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /{short_url} [get]
func (h *URLsHandler) GetURL(c *gin.Context) {
	shortURL := c.Param("short_url")
	h.logger.Info("GetURL called", zap.String("short_url", shortURL))

	// TODO: Implement actual URL retrieval logic
	c.JSON(http.StatusPermanentRedirect, gin.H{"url": "https://example.com"})
}
