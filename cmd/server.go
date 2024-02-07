package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config *OrcaConfig
	router *gin.Engine
}

func NewServer(config *OrcaConfig) *Server {
	return &Server{
		config: config,
		router: NewRouter(config),
	}
}

func NewRouter(config *OrcaConfig) *gin.Engine {
	gin.ForceConsoleColor()

	router := gin.New()

	var logWriter io.Writer
	if config.Loglevel != "debug" {
		logWriter = *NewMaskedWriter()
	}
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: logWriter}), gin.Recovery())

	return router
}

func (s *Server) AddRoutes() {
	// health
	s.router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	authorized := s.router.Group("/", gin.BasicAuth(s.config.Basicauth))

	// autosync
	authorized.POST("/sync/:mode", func(c *gin.Context) {
		switch c.Param("mode") {
		case "on", "off":
			s.config.Autosync = c.Param("mode")
			c.String(http.StatusOK, "toggled")
		default:
			c.String(http.StatusBadRequest, "mode must be one of 'on' or 'off'")
		}

	})

}
