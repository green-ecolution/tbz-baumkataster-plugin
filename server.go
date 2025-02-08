package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/green-ecolution/green-ecolution-backend/pkg/plugin"
)

type ServerConfig struct {
	port     int
	plugin   plugin.Plugin
	pluginFS fs.FS
	version  string
}

type Server struct {
	cfg *ServerConfig
}

type ServerOption func(*ServerConfig)

func WithPort(port int) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.port = port
	}
}

func WithVersion(version string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.version = version
	}
}

func WithPluginFS(pluginFS fs.FS) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.pluginFS = pluginFS
	}
}

func WithPlugin(plugin plugin.Plugin) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.plugin = plugin
	}
}

var defaultServerConfig = &ServerConfig{
	port:    8080,
	version: "develop",
}

func NewServer(opts ...ServerOption) *Server {
	cfg := defaultServerConfig
	for _, opt := range opts {
		opt(cfg)
	}
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) Run(ctx context.Context) error {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServerFS(s.cfg.pluginFS))
		fs.ServeHTTP(w, r)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.port),
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(timeoutCtx); err != nil {
			slog.Error("failed to shutdown http server", "error", err)
		}
	}()

	return server.ListenAndServe()
}
