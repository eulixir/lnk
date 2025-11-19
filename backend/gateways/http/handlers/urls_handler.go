package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"lnk/domain/entities/usecases"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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

// CreateURL creates a short URL from a long URL.
//
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
	ctx := c.Request.Context()
	tracer := otel.Tracer("handlers.CreateURL")
	ctx, span := tracer.Start(ctx, "CreateURLHandler")

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()
	defer span.End()

	var req CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = fmt.Errorf("failed to bind JSON: %w", err)
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	shortURL, err := h.useCase.CreateShortURL(ctx, req.URL)
	if err != nil {
		err = fmt.Errorf("failed to create short URL: %w", err)
		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	span.SetStatus(codes.Ok, "Short URL created")
	c.JSON(http.StatusOK, CreateURLResponse{
		ShortURL:    shortURL,
		OriginalURL: req.URL,
	})
}

// GetURL retrieves the original URL from a short URL.
//
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
	ctx := c.Request.Context()
	tracer := otel.Tracer("handlers.GetURL")
	ctx, span := tracer.Start(ctx, "GetURLHandler")

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()
	defer span.End()

	longURL, err := h.useCase.GetLongURL(ctx, shortCode)
	if err != nil {
		if errors.Is(err, usecases.ErrURLNotFound) {
			span.SetStatus(codes.Error, err.Error())
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}

		span.SetStatus(codes.Error, err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})

		return
	}

	span.SetStatus(codes.Ok, "URL found")
	c.JSON(http.StatusPermanentRedirect, gin.H{"url": longURL})
}
