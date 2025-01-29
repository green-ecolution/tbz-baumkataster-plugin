package main

import (
	"context"
	"fmt"

	"github.com/green-ecolution/green-ecolution-backend/client"
)

type GreenEcolutionClient struct {
	client *client.APIClient
}

func NewGreenEcolutionRepo(cfg *client.Configuration) *GreenEcolutionClient {
	return &GreenEcolutionClient{
		client: client.NewAPIClient(cfg),
	}
}

func (r *GreenEcolutionClient) GetInfo(ctx context.Context) (*client.AppInfo, error) {
	info, _, err := r.client.InfoAPI.GetAppInfo(ctx).Execute()
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (r *GreenEcolutionClient) Create(ctx context.Context, tree Tree) (int32, error) {
	body := client.TreeCreate{
		Description:  fmt.Sprintf("%s - Dieser Baum wurde importiert", tree.Description),
		Latitude:     float32(tree.Latitude),
		Longitude:    float32(tree.Longitude),
		Number:       tree.TreeNumber,
		PlantingYear: tree.PlantingYear,
		Readonly:     true,
		Species:      tree.Species,
	}

	respTree, _, err := r.client.TreeAPI.CreateTree(ctx).Body(body).Execute()
	if err != nil {
		return 0, err
	}

	return respTree.Id, nil
}

// func (r *GreenEcolutionClient) Update(ctx context.Context, id int, tree Tree) (Tree, error) {
//
// }
//
// func (r *GreenEcolutionClient) Archive(ctx context.Context, id int) error {
//
// }
