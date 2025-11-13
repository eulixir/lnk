package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title           LNK URL Shortener API
// @version         1.0
// @description     A URL shortener service API
// @termsOfService  http://swagger.io/terms/

// @BasePath  /

// @schemes   http https

type Server struct {
	logger *zap.Logger
	port   string
	srv    *http.Server
	router *gin.Engine
}

type Config struct {
	Logger *zap.Logger
	Port   string
	Router *gin.Engine
}

func NewServer(logger *zap.Logger, port string, router *gin.Engine) *Server {
	return &Server{
		logger: logger,
		port:   port,
		router: router,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
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
