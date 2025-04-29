package main

import (
	"slices"
	"strconv"
	"strings"

	"github.com/green-ecolution/backend/pkg/client"
)

type Tree struct {
	client.Tree
	TreeRegisterID int
}

type TreeRegister struct {
	ID           int    `db:"OBJECTID"`
	TreeNumber   string `db:"BAUMNUMMER"`
	Hochwert     string `db:"HOCHWERT"`
	Rechtswert   string `db:"RECHTSWERT"`
	Species      string `db:"GATTUNG"`
	Region       string `db:"GEBIET"`
	Street       string `db:"STRASSE"`
	PlantingYear int    `db:"PFLANZJAHR"`
}

const (
	fromEPSG = "EPSG:31467"
	toEPSG   = "EPSG:4326"
)

func TreesFromBatch(register []TreeRegister) ([]Tree, error) {
	transformer, err := NewGeoTransformer(fromEPSG, toEPSG)
	if err != nil {
		return nil, err
	}
	defer transformer.Destroy()

	geoPointsSeq := MapIter12(slices.Values(register), func(r TreeRegister) (GeoPoint, error) {
		rCoord, err := strconv.ParseFloat(strings.Replace(r.Rechtswert, ",", ".", 1), 64)
		if err != nil {
			return GeoPoint{}, err
		}

		hCoord, err := strconv.ParseFloat(strings.Replace(r.Hochwert, ",", ".", 1), 64)
		if err != nil {
			return GeoPoint{}, err
		}

		return GeoPoint{
			X: hCoord,
			Y: rCoord,
		}, nil
	})

	geoPoints, err := CollectOrError(geoPointsSeq)
	if err != nil {
		return nil, err
	}

	geoPoints, err = transformer.TransformBatch(geoPoints)
	if err != nil {
		return nil, err
	}

	treeSeq := MapIter21(Zip(slices.Values(register), slices.Values(geoPoints)), func(r TreeRegister, g GeoPoint) Tree {
		return Tree{
			TreeRegisterID: r.ID,
			Tree: client.Tree{
				PlantingYear: int32(r.PlantingYear),
				Species:      r.Species,
				Number:       r.TreeNumber,
				Description:  r.Region,
				Latitude:     float32(g.X),
				Longitude:    float32(g.Y),
			},
		}
	})

	return slices.Collect(treeSeq), nil
}
