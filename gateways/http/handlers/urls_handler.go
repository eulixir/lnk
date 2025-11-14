package handlers

import (
	"errors"
	"net/http"

	"lnk/domain/entities/usecases"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CreateURLRequest struct {
	URL string `json:"url" example:"https://example.com" binding:"required"`
}

type CreateURLResponse struct {
	ShortURL    string `json:"short_url" example:"abc123"`
	OriginalURL string `json:"original_url" example:"https://example.com"`
}

type GetURLResponse struct {
	ShortURL    string `json:"short_url" example:"abc123"`
	OriginalURL string `json:"original_url" example:"https://example.com"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type URLsHandler struct {
	logger  *zap.Logger
	useCase *usecases.UseCase
}

func NewURLsHandler(logger *zap.Logger, useCase *usecases.UseCase) *URLsHandler {
	return &URLsHandler{
		logger:  logger,
		useCase: useCase,
	}
}

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

	shortURL, err := h.useCase.CreateShortURL(c.Request.Context(), req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, CreateURLResponse{
		ShortURL:    shortURL,
		OriginalURL: req.URL,
	})
}

// @Summary      Get original URL by short URL
// @Description  Get the original URL from a short URL
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        short_url  path      string  true  "Short URL identifier"
// @Success      308        {object}  map[string]string
// @Failure      404        {object}  ErrorResponse
// @Failure      500        {object}  map[string]string
// @Router       /{short_url} [get]
func (h *URLsHandler) GetURL(c *gin.Context) {
	shortCode := c.Param("short_url")

	longURL, err := h.useCase.GetLongURL(shortCode)
	if err != nil {
		if errors.Is(err, usecases.ErrURLNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusPermanentRedirect, gin.H{"url": longURL})
}
