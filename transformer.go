package main

import (
	"github.com/twpayne/go-proj/v10"
)

type GeoTransformer struct {
	pj *proj.PJ
}

func NewGeoTransformer(from, to string) (*GeoTransformer, error) {
	pj, err := proj.NewCRSToCRS(from, to, nil)
	if err != nil {
		return nil, err
	}

	return &GeoTransformer{
		pj: pj,
	}, nil
}

func (g *GeoTransformer) Transform(x, y float64) (lat, lng float64, err error) {
	coord, err := g.pj.Forward(proj.NewCoord(x, y, 0, 0))
	if err != nil {
		return 0, 0, err
	}

	return coord.X(), coord.Y(), nil
}

type GeoPoint struct {
	X float64
	Y float64
}

func (g *GeoTransformer) TransformBatch(points []GeoPoint) ([]GeoPoint, error) {
	coords := Map(points, func(p GeoPoint) proj.Coord {
		return proj.NewCoord(p.X, p.Y, 0, 0)
	})

	if err := g.pj.ForwardArray(coords); err != nil {
		return nil, err
	}

	return Map(coords, func(c proj.Coord) GeoPoint {
		return GeoPoint{X: c.X(), Y: c.Y()}
	}), nil
}

func (g *GeoTransformer) Destroy() {
	g.pj.Destroy()
}
