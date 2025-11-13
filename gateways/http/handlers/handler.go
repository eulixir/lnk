package handlers

import (
	"net/http"

	"lnk/domain/entities/usecases"
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
	useCase     *usecases.UseCase
}

type HttpHandlers struct {
	Router  *gin.Engine
	Logger  *zap.Logger
	Env     string
	UseCase *usecases.UseCase
}

func NewHttpHandlers(httpHandlers *HttpHandlers) *Handlers {
	h := &Handlers{
		router:      httpHandlers.Router,
		logger:      httpHandlers.Logger,
		env:         httpHandlers.Env,
		URLsHandler: NewURLsHandler(httpHandlers.Logger),
		useCase:     httpHandlers.UseCase,
	}

	h.setupMiddleware()
	h.setupHandlers()

	return h
}

func (h *Handlers) setupMiddleware() {
	h.router.Use(middleware.Recovery(h.logger))

	h.router.Use(middleware.RequestLogger(h.logger))

	h.router.Use(middleware.CORS())
}

func (h *Handlers) setupHandlers() {
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
