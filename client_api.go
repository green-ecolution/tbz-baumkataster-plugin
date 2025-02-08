package main

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/green-ecolution/green-ecolution-backend/pkg/client"
)

type GreenEcolutionClient struct {
	client   *client.APIClient
	provider string
}

func NewGreenEcolutionRepo(cfg *client.Configuration, provider string) *GreenEcolutionClient {
	return &GreenEcolutionClient{
		client:   client.NewAPIClient(cfg),
		provider: provider,
	}
}

func (r *GreenEcolutionClient) GetInfo(ctx context.Context) (*client.AppInfo, error) {
	info, _, err := r.client.InfoAPI.GetAppInfo(ctx).Execute()
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (r *GreenEcolutionClient) GetAll(ctx context.Context) ([]Tree, error) {
	trees, _, err := r.client.TreeAPI.GetAllTrees(ctx).Provider(r.provider).Execute()
	if err != nil {
		return nil, err
	}

	toReset := FilterIter(slices.Values(trees.Data), func(t client.Tree) bool {
		v, ok := t.AdditionalInformation["object_id"]
		if !ok {
			return true
		}

		v, ok = v.(float64)
		if !ok {
			return true
		}

		if v == 0 {
			return true
		}

		return false
	})

	// Delete/Reset Trees if there has no object id in additional info field
	for t := range toReset {
		if err := r.Delete(ctx, t.Id); err != nil {
			slog.Warn("failed to delete tree in green ecolution backend", "error", err, "tree_id", t.Id)
		}
	}

	filteredClientTrees := FilterIter(slices.Values(trees.Data), func(t client.Tree) bool {
		v, ok := t.AdditionalInformation["object_id"]
		if !ok {
			return false
		}

		v, ok = v.(float64)
		if !ok || v == 0 {
			return false
		}

		return true
	})

	mapSeq := MapIter(filteredClientTrees, func(t client.Tree) Tree {
		return Tree{
			Tree:           t,
			TreeRegisterID: int(t.AdditionalInformation["object_id"].(float64)),
		}
	})

	return slices.Collect(mapSeq), nil
}

func (r *GreenEcolutionClient) Create(ctx context.Context, tree Tree) error {
	body := client.TreeCreate{
		Description:  fmt.Sprintf("%s - Dieser Baum wurde importiert", tree.Description),
		Latitude:     float32(tree.Latitude),
		Longitude:    float32(tree.Longitude),
		Number:       tree.Number,
		PlantingYear: tree.PlantingYear,
		Readonly:     true,
		Species:      tree.Species,
		Provider:     &r.provider,
		AdditionalInformation: map[string]interface{}{
			"object_id": tree.TreeRegisterID,
		},
	}

	_, _, err := r.client.TreeAPI.CreateTree(ctx).Body(body).Execute()
	return err
}

func (r *GreenEcolutionClient) Update(ctx context.Context, id int32, tree Tree) error {
	body := client.TreeUpdate{
		Description:  tree.Description,
		Latitude:     float32(tree.Latitude),
		Longitude:    float32(tree.Longitude),
		Number:       tree.Number,
		PlantingYear: tree.PlantingYear,
		Readonly:     true,
		Species:      tree.Species,
		Provider:     &r.provider,
		AdditionalInformation: map[string]interface{}{
			"object_id": tree.TreeRegisterID,
		},
	}

	_, _, err := r.client.TreeAPI.UpdateTree(ctx, id).Body(body).Execute()
	return err
}

func (r *GreenEcolutionClient) Delete(ctx context.Context, id int32) error {
	_, err := r.client.TreeAPI.DeleteTree(ctx, id).Execute()
	return err
}

func (r *GreenEcolutionClient) Archive(ctx context.Context, id int32) error {
	_, err := r.client.TreeAPI.DeleteTree(ctx, id).Execute() // TODO: Archive
	return err
}
