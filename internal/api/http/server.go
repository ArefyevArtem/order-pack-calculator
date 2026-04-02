package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"order-pack-calculator/internal/api/http/middlewares"
)

// ServerConfig holds HTTP listen options.
type ServerConfig struct {
	Host         string        `envconfig:"HOST" default:"0.0.0.0"`
	Port         string        `envconfig:"PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
}

// Addr returns host:port for the HTTP server.
func (c ServerConfig) Addr() string {
	return c.Host + ":" + c.Port
}

// Controller registers its routes on the Gin engine.
type Controller interface {
	RegisterRoutes(r *gin.Engine)
}

// Server is the API HTTP server (Gin + middlewares + controllers).
type Server struct {
	cfg         ServerConfig
	controllers []Controller
	srv         *http.Server
}

// NewServer builds an empty server; add controllers via AddController.
func NewServer(cfg ServerConfig) *Server {
	return &Server{cfg: cfg, controllers: nil}
}

// AddController appends one or more controllers.
func (s *Server) AddController(c ...Controller) {
	s.controllers = append(s.controllers, c...)
}

// Start wires Gin, runs ListenAndServe in a goroutine, blocks until ctx is cancelled, then shuts down gracefully.
func (s *Server) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.RequestLogger)

	for _, c := range s.controllers {
		c.RegisterRoutes(r)
	}

	s.srv = &http.Server{
		Addr:         s.cfg.Addr(),
		Handler:      r,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
