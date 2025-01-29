package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"

	"github.com/green-ecolution/green-ecolution-backend/client"
	"github.com/green-ecolution/green-ecolution-backend/plugin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

const (
	slug = "tbz-baumkataster"
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
		plugin.WithSlug(slug),
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

	geClient := NewGreenEcolutionRepo(clientCfg, slug)

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

	mapRegisterTrees, err := TreesFromBatch(registerTrees)
	if err != nil {
		panic(err)
	}

	geTrees, err := geClient.GetAll(ctx)
	if err != nil {
		panic(err)
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
			if updatedTree, ok := CheckDiff(regTree, geTree); !ok {
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
		if err := geClient.Create(ctx, e); err != nil {
			slog.Warn("failed to create tree in green ecolution backend", "error", err, "register_id", e.TreeRegisterID)
		}
	}

	for _, e := range updateQueue {
		if err := geClient.Update(ctx, e.Id, e); err != nil {
			slog.Warn("failed to update tree in green ecolution backend", "error", err, "register_id", e.TreeRegisterID, "tree_id", e.Id)
		}
	}

	for _, e := range archiveQueue {
		if err := geClient.Archive(ctx, e.Id); err != nil {
			slog.Warn("failed to archive tree in green ecolution backend", "error", err, "register_id", e.TreeRegisterID, "tree_id", e.Id)
		}
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

func CheckDiff(new, old Tree) (Tree, bool) {
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
