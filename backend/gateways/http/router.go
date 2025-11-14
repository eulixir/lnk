package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	_ "lnk/docs"
	"lnk/gateways/http/handlers"
	"lnk/gateways/http/middleware"
)

type RouterConfig struct {
	Logger   *zap.Logger
	GinMode  string
	Env      string
	Handlers *handlers.Handlers
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	router := gin.New()

	router.Use(middleware.Recovery(cfg.Logger))
	router.Use(middleware.RequestLogger(cfg.Logger))
	router.Use(middleware.CORS())

	cfg.Handlers.RegisterRoutes(router, cfg.Env)

	return router
}
