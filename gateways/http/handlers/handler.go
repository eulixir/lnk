package handlers

import (
	"net/http"

	"lnk/gateways/http/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Handlers struct {
	router      *gin.Engine
	logger      *zap.Logger
	env         string
	URLsHandler *URLsHandler
}

func NewHttpHandlers(router *gin.Engine, logger *zap.Logger, env string) *Handlers {
	h := &Handlers{
		router:      router,
		logger:      logger,
		env:         env,
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
	if h.env == "development" {
		h.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	h.router.GET("/health", h.healthCheck)
	h.router.POST("/shorten", h.URLsHandler.CreateURL)
	h.router.GET("/:short_url", h.URLsHandler.GetURL)
}

// healthCheck godoc
// @Summary      Health check endpoint
// @Description  Check if the API is running
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *Handlers) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
