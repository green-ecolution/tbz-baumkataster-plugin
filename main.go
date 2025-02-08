package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"slices"
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
		plugin.WithSlug(slug),
		plugin.WithVersion("v0.0.1"),
		plugin.WithHostPath(pluginPath),
	) // TODO: change to no ptr in backend

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
	worker.SetClient(oauthClient)

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

type SyncTrees struct {
	registerRepo *TreeRegisterRepo
	client       *GreenEcolutionClient
}

func NewSyncTrees(repo *TreeRegisterRepo, client *GreenEcolutionClient) *SyncTrees {
	return &SyncTrees{
		registerRepo: repo,
		client:       client,
	}
}

func (s *SyncTrees) Sync(ctx context.Context) error {
	slog.Info("sync tbz register trees to green ecolution backend")
	registerTrees, err := s.registerRepo.GetTreesPlantedAfter(ctx)
	if err != nil {
		slog.Error("failed to get trees from tbz tree register", "error", err)
		return nil
	}

	mapRegisterTrees, err := TreesFromBatch(registerTrees)
	if err != nil {
		slog.Error("failed to map tbz tree register to internal trees", "error", err)
		return nil
	}

	geTrees, err := s.client.GetAll(ctx)
	if err != nil {
		slog.Error("failed to get trees from green ecolution backend", "error", err)
		return nil
	}

	slices.SortFunc(mapRegisterTrees, func(a Tree, b Tree) int {
		return a.TreeRegisterID - b.TreeRegisterID
	})

	slices.SortFunc(geTrees, func(a Tree, b Tree) int {
		return a.TreeRegisterID - b.TreeRegisterID
	})

	createdQueue := make([]Tree, 0)
	updateQueue := make([]Tree, 0)
	archiveQueue := make([]Tree, 0)

	idxRegTrees := 0
	idxGeTrees := 0

	for idxRegTrees < len(mapRegisterTrees) || idxGeTrees < len(geTrees) {
		if idxRegTrees == len(mapRegisterTrees) {
			archiveQueue = append(archiveQueue, geTrees[idxRegTrees:]...)
			break
		}

		if idxGeTrees == len(geTrees) {
			createdQueue = append(createdQueue, mapRegisterTrees[idxGeTrees:]...)
			break
		}

		regTree := mapRegisterTrees[idxRegTrees]
		geTree := geTrees[idxGeTrees]

		if regTree.TreeRegisterID == geTree.TreeRegisterID {
			if updatedTree, ok := s.checkDiff(regTree, geTree); !ok {
				updateQueue = append(updateQueue, updatedTree)
			}
			idxGeTrees++
			idxRegTrees++
			continue
		}

		if regTree.TreeRegisterID < geTree.TreeRegisterID {
			createdQueue = append(createdQueue, regTree)
			idxRegTrees++
			continue
		}

		if regTree.TreeRegisterID > geTree.TreeRegisterID {
			archiveQueue = append(archiveQueue, geTree)
			idxGeTrees++
			continue
		}
	}

	for _, e := range createdQueue {
		if err := s.client.Create(ctx, e); err != nil {
			slog.Warn("failed to create tree in green ecolution backend", "error", err, "register_id", e.TreeRegisterID)
		}
	}

	for _, e := range updateQueue {
		if err := s.client.Update(ctx, e.Id, e); err != nil {
			slog.Warn("failed to update tree in green ecolution backend", "error", err, "register_id", e.TreeRegisterID, "tree_id", e.Id)
		}
	}

	for _, e := range archiveQueue {
		if err := s.client.Archive(ctx, e.Id); err != nil {
			slog.Warn("failed to archive tree in green ecolution backend", "error", err, "register_id", e.TreeRegisterID, "tree_id", e.Id)
		}
	}

	return nil
}

func (s *SyncTrees) checkDiff(new, old Tree) (Tree, bool) {
	if new.Number == old.Number ||
		new.Latitude == old.Latitude ||
		new.Longitude == old.Longitude ||
		new.Species == old.Species ||
		new.PlantingYear == old.PlantingYear {
		return new, true
	} else {
		old.Number = new.Number
		old.Latitude = new.Latitude
		old.Longitude = new.Longitude
		old.Species = new.Species
		old.PlantingYear = new.PlantingYear

		return old, false
	}
}
