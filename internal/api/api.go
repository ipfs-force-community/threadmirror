package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	v1 "github.com/ipfs-force-community/threadmirror/internal/api/v1"
	v1middleware "github.com/ipfs-force-community/threadmirror/internal/api/v1/middleware"
	"github.com/ipfs-force-community/threadmirror/internal/config"
	"github.com/ipfs-force-community/threadmirror/pkg/auth"
	"github.com/ipfs-force-community/threadmirror/pkg/i18n"
	sloggin "github.com/samber/slog-gin"
)

type Server struct {
	GinEngine  *gin.Engine
	HttpServer *http.Server
	logger     *slog.Logger
}

func NewServer(
	logger *slog.Logger,
	serverConfig *config.ServerConfig,
	v1Handler *v1.V1Handler,
	i18nBundle *i18n.I18nBundle,
	jwtVerifier auth.JWTVerifier,
) *Server {
	if !serverConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	engine.Use(sloggin.New(logger))
	engine.Use(gin.Recovery())
	engine.Use(i18n.Middleware(i18nBundle))
	engine.Use(v1middleware.ErrorHandler())

	if serverConfig.Debug {
		// config CORS，always allow cross-origin requests
		c := cors.DefaultConfig()
		c.AllowAllOrigins = true
		c.AllowHeaders = append(c.AllowHeaders, "Authorization")
		c.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		engine.Use(cors.New(c))
	}

	v1.RegisterHandlersWithOptions(
		engine,
		v1Handler,
		v1.GinServerOptions{
			BaseURL: "/api/v1",
			Middlewares: []v1.MiddlewareFunc{
				v1middleware.Authentication(
					auth.Middleware(jwtVerifier, func(c *gin.Context, statusCode int) {
						c.AbortWithStatusJSON(statusCode, gin.H{
							"error": v1.T(c, "ErrorAuthenticationRequired"),
						})
					}),
				),
			},
		},
	)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         serverConfig.Addr,
		Handler:      engine,
		ReadTimeout:  serverConfig.ReadTimeout,
		WriteTimeout: serverConfig.WriteTimeout,
	}

	return &Server{
		logger:     logger,
		GinEngine:  engine,
		HttpServer: httpServer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP server", "address", s.HttpServer.Addr)
	s.HttpServer.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	addr := s.HttpServer.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("start http server listen on %s: %w", addr, err)
	}
	go func() {
		err = s.HttpServer.Serve(ln)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.HttpServer.Shutdown(ctx)
}
