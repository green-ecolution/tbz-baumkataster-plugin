package main

import (
	"slices"
	"strconv"
	"strings"
)

type Tree struct {
	TreeRegisterID       string  `db:"register_id"`
	GreenEcolutionTreeID int32   `db:"green_ecolution_id"`
	PlantingYear         int32   `db:"planting_year"`
	Species              string  `db:"species"`
	TreeNumber           string  `db:"tree_number"`
	Latitude             float64 `db:"latitude"`
	Longitude            float64 `db:"longitude"`
	Description          string  `db:"description"`
}

type TreeRegister struct {
	ID           string `db:"OBJECTID"`
	TreeNumber   string `db:"BAUMNUMMER"`
	Hochwert     string `db:"HOCHWERT"`
	Rechtswert   string `db:"RECHTSWERT"`
	Species      string `db:"GATTUNG"`
	Region       string `db:"GEBIET"`
	Street       string `db:"STRASSE"`
	PlantingYear int    `db:"PFLANZJAHR"`
}

const (
	fromEPSG = 31467
	toEPSG   = 4326
)

func TreesFromBatch(register []TreeRegister) ([]Tree, error) {
	transformer, err := NewGeoTransformer(fromEPSG, toEPSG)
	if err != nil {
		return nil, err
	}

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
			PlantingYear:   int32(r.PlantingYear),
			Species:        r.Species,
			TreeNumber:     r.TreeNumber,
			Description:    r.Region,
			Latitude:       g.X,
			Longitude:      g.Y,
		}
	})

	return slices.Collect(treeSeq), nil
}
