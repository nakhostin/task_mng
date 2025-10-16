package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"task_mng/cmd/web/config"
	userR "task_mng/domain/user"
	"task_mng/interfaces/http/handlers"
	"task_mng/interfaces/http/middleware"
	"task_mng/pkg/jwt"
	"task_mng/pkg/postgres"
	"task_mng/pkg/redis"
	"task_mng/services/user"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	config   *config.Config
	router   *gin.Engine
	server   *http.Server
	jwtMng   *jwt.Manager
	postgres *postgres.Database
	redis    *redis.Redis
	handlers *handlers.Handlers
}

func New(config *config.Config, jwtMng *jwt.Manager, postgres *postgres.Database, redis *redis.Redis) *Server {
	router := gin.Default()

	// Set max body size to 3MB
	router.MaxMultipartMemory = 3 << 20 // 3MB

	userRepo := userR.New(postgres)
	userService := user.New(userRepo, jwtMng)

	srv := &Server{
		config:   config,
		router:   router,
		jwtMng:   jwtMng,
		postgres: postgres,
		redis:    redis,
		handlers: handlers.New(userService),
	}

	srv.setupRoutes()

	return srv
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	slog.Info("Starting HTTP server", "address", addr)

	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down HTTP server")
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// Router returns the Gin router instance
func (s *Server) Router() *gin.Engine {
	return s.router
}

// setupRoutes sets up the server routes
func (s *Server) setupRoutes() {
	// Swagger documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := s.router.Group("/api/v1")

	// ********************* Auth routes *********************
	auth := v1.Group("/auth")
	auth.POST("/login", s.handlers.User.Login)
	auth.POST("/refresh", s.handlers.User.Refresh)

	protected := v1.Group("")
	protected.Use(middleware.LoginRequired(s.jwtMng))

	user := protected.Group("/users")
	user.POST("", s.handlers.User.Create)
	user.GET("", s.handlers.User.FindAll)

	profile := protected.Group("/profile")
	profile.GET("", s.handlers.User.Me)
	profile.PUT("", s.handlers.User.Update)
}
