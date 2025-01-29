package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/green-ecolution/green-ecolution-backend/client"
	"github.com/green-ecolution/green-ecolution-backend/plugin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Error loading .env file")
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
		plugin.WithSlug("tbz-baumkataster"),
		plugin.WithVersion("v0.0.1"),
		plugin.WithHostPath(pluginPath),
	) // TODO: change to no ptr in backend

	worker, err := plugin.NewPluginWorker(
		plugin.WithHost(hostPath),
		plugin.WithPlugin(*p),
		plugin.WithHostAPIVersion("v1"),
	)
	if err != nil {
		panic(err)
	}

	token, err := worker.Register(ctx, clientID, clientSecret)
	if err != nil {
		panic(err)
	}
	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		// ExpiresIn:    token.ExpiresIn,
		TokenType: "Bearer",
	}
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(oauthToken))
	clientCfg := client.NewConfiguration()
	clientCfg.Servers = client.ServerConfigurations{
		{
			URL:         fmt.Sprintf("%s/api", hostPath),
			Description: "Green Ecolution API",
		},
	}
	clientCfg.Debug = true
	clientCfg.HTTPClient = oauthClient

	geClient := NewGreenEcolutionRepo(clientCfg)

	auth := context.WithValue(ctx, client.ContextOAuth2, oauthToken)

	info, err := geClient.GetInfo(auth)
	if err != nil {
		slog.Error("Error while getting app info", "error", err)
	}
	slog.Info("App info", "info", info)

	dsn := os.Getenv("DB_URL")
	repo, err := NewTreeRegisterRepo(dsn)
	if err != nil {
		panic(err)
	}

	registerTrees, err := repo.GetTreesPlantedAfter(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(registerTrees)

	localStorage, err := NewLocalStorageRepo("db.sqlite3")
	if err != nil {
		panic(err)
	}

	if err := localStorage.Setup(ctx); err != nil {
		panic(err)
	}

	mapTrees, err := TreesFromBatch(registerTrees)
	if err != nil {
		panic(err)
	}

	if err := localStorage.Insert(ctx, mapTrees); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := worker.RunHeartbeat(ctx); err != nil {
			slog.Error("Failed to send heartbeat", "error", err)
		}
	}()

	wg.Wait()
}
