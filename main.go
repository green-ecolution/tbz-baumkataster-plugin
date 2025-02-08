package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/green-ecolution/green-ecolution-backend/pkg/client"
	"github.com/green-ecolution/green-ecolution-backend/pkg/plugin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

const (
	slug = "tbz-baumkataster"
)

var version = "develop"

//go:embed all:ui/dist
var f embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("error loading .env file")
	}

	clientSecret := os.Getenv("CLIENT_SECRET")
	clientID := os.Getenv("CLIENT_ID")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	hostPath, err := url.Parse("http://localhost:3000")
	if err != nil {
		panic(err)
	}

	pluginPath, err := url.Parse("http://localhost:6123")
	if err != nil {
		panic(err)
	}

	p := plugin.NewPlugin(
		plugin.WithName("TBZ Baumkataster"),
		plugin.WithDescription("Dieses Plugin ist für die Synchronisation der Bäume im TBZ Baumkataster mit den Bäumen im Green Ecolution System zuständig. Dabei wird in regelmäßigen Abständen geprüft, ob neue Bäume im Baumkataster hinzugefügt, angepasst oder entfernt wurden. Dabei werden nur Bäume innerhalb eines definierten Standjahres von bis zu drei Jahren synchronisiert."),
		plugin.WithSlug(slug),
		plugin.WithVersion(version),
		plugin.WithHostPath(pluginPath),
	)

	worker, err := plugin.NewPluginWorker(
		plugin.WithHost(hostPath),
		plugin.WithPlugin(p),
		plugin.WithHostAPIVersion("v1"),
	)
	if err != nil {
		panic(err)
	}

	token, err := worker.Register(ctx, clientID, clientSecret)
	if err != nil {
		panic(err)
	}

	oauthClient := authClient(ctx, token)
	worker.SetClient(oauthClient)

	clientCfg := client.NewConfiguration()
	clientCfg.Servers = client.ServerConfigurations{
		{
			URL:         fmt.Sprintf("%s/api", hostPath),
			Description: "Green Ecolution API",
		},
	}
	clientCfg.Debug = true
	clientCfg.HTTPClient = oauthClient

	geClient := NewGreenEcolutionRepo(clientCfg, slug)

	dsn := os.Getenv("DB_URL")
	repo, err := NewTreeRegisterRepo(dsn)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(4)

	fSub, err := fs.Sub(f, "ui/dist")
	if err != nil {
		panic(err)
	}

	server := NewServer(
		WithPort(6123),
		WithPlugin(p),
		WithPluginFS(fSub),
		WithVersion(version),
	)

	go func() {
		defer wg.Done()
		if err := server.Run(ctx); err != nil {
			slog.Error("failed to start http server", "error", err)
		}
	}()

	syncTrees := NewSyncTrees(repo, geClient)
	scheduler := NewScheduler(10 * time.Second)
	go func() {
		defer wg.Done()
		if err := scheduler.Run(ctx, syncTrees.Sync); err != nil {
			slog.Error("an error has occurred in syncing trees from tbz tree register to green ecolution backend", "error", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := worker.RunHeartbeat(ctx); err != nil {
			slog.Error("Failed to send heartbeat", "error", err)
		}
	}()

	go func() {
		defer wg.Done()
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := worker.Unregister(timeoutCtx); err != nil {
			slog.Error("failed to unregister plugin", "error", err)
		}
	}()

	wg.Wait()
}

func authClient(ctx context.Context, token *plugin.Token) *http.Client {
	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    "Bearer",
	}

	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(oauthToken))
}
