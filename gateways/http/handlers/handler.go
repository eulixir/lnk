package handlers

import (
	"net/http"

	"lnk/gateways/http/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handlers struct {
	router      *gin.Engine
	logger      *zap.Logger
	URLsHandler *URLsHandler
}

func NewHttpHandlers(router *gin.Engine, logger *zap.Logger) *Handlers {
	h := &Handlers{
		router:      router,
		logger:      logger,
		URLsHandler: NewURLsHandler(logger),
	}

	h.setupMiddleware()

	return h
}

func (h *Handlers) setupMiddleware() {
	h.router.Use(middleware.Recovery(h.logger))

	h.router.Use(middleware.RequestLogger(h.logger))

	h.router.Use(middleware.CORS())
}

func (h *Handlers) SetupHandlers() {
	h.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	h.router.POST("/shorten", h.URLsHandler.CreateURL)
	h.router.GET("/:short_url", h.URLsHandler.GetURL)
}
