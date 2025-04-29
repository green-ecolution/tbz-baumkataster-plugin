package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/green-ecolution/backend/pkg/plugin"
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

	r.Get("/", s.handleHelloWorld)
	r.Get("/info", s.handleGetInfo)
	r.Post("/sync", s.handleExecSync)
	r.Post("/reset", s.handleExecReset)

	r.Get("/*", s.handleFileSystem)

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

func (s *Server) handleHelloWorld(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello World"))
}

type info struct {
	SyncInterval      string    `json:"sync_interval"`
	LastSync          time.Time `json:"last_sync"`
	PluginSlug        string    `json:"slug"`
	Version           string    `json:"version"`
	TotalManagedTrees int       `json:"total_managed_trees"`
	Description       string    `json:"description"`
}

func (s *Server) handleGetInfo(w http.ResponseWriter, _ *http.Request) {
	cfg := ParseConfig()
	infoResp := info{
		SyncInterval:      cfg.SyncInterval.String(),
		LastSync:          syncTrees.lastSync,
		PluginSlug:        cfg.PluginSlug,
		Version:           version,
		TotalManagedTrees: syncTrees.managedTrees,
		Description:       description,
	}

	encode := json.NewEncoder(w)
	encode.Encode(infoResp)
}

func (s *Server) handleExecSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := syncTrees.Sync(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleExecReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := syncTrees.Reset(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleFileSystem(w http.ResponseWriter, r *http.Request) {
	rctx := chi.RouteContext(r.Context())
	pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
	fs := http.StripPrefix(pathPrefix, http.FileServerFS(s.cfg.pluginFS))
	fs.ServeHTTP(w, r)
}
