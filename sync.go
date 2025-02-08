package main

import (
	"context"
	"log/slog"
	"slices"
	"time"
)

type SyncTrees struct {
	registerRepo *TreeRegisterRepo
	client       *GreenEcolutionClient
	lastSync     time.Time
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

	s.lastSync = time.Now()
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
