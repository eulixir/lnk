package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"lnk/domain/entities/usecases"
)

type Handlers struct {
	logger      *zap.Logger
	URLsHandler *URLsHandler
	useCase     *usecases.UseCase
}

func NewHandlers(logger *zap.Logger, useCase *usecases.UseCase) *Handlers {
	return &Handlers{
		logger:      logger,
		URLsHandler: NewURLsHandler(logger, useCase),
		useCase:     useCase,
	}
}

func (h *Handlers) RegisterRoutes(router *gin.Engine, env string) {
	if env == "development" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	router.GET("/health", h.healthCheck)

	router.POST("/shorten", h.URLsHandler.CreateURL)
	router.GET("/:short_url", h.URLsHandler.GetURL)
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
