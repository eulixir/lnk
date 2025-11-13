package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"lnk/extensions/config"
	"lnk/gateways/http/handlers"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// @title           LNK URL Shortener API
// @version         1.0
// @description     A URL shortener service API
// @termsOfService  http://swagger.io/terms/

// @BasePath  /

// @schemes   http https

type Server struct {
	router   *gin.Engine
	logger   *zap.Logger
	config   *config.Config
	handlers *handlers.Handlers
	srv      *http.Server
}

func NewServer(logger *zap.Logger, cfg *config.Config) *Server {
	router := gin.Default()

	httpHandlers := handlers.NewHttpHandlers(router, logger)

	return &Server{
		router:   router,
		logger:   logger,
		config:   cfg,
		handlers: httpHandlers,
	}
}

func (s *Server) Start() error {
	s.handlers.SetupHandlers()

	addr := fmt.Sprintf(":%s", s.config.Port)
	s.logger.Info("Starting HTTP server", zap.String("address", addr))

	s.srv = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}

	s.logger.Info("Shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("HTTP server stopped")
	return nil
}
