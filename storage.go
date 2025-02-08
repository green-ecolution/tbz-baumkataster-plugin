package main

import (
	"context"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const query = `
select "OBJECTID", "BAUMNUMMER", "HOCHWERT", "RECHTSWERT", "GATTUNG", "GEBIET", "STRASSE", "PFLANZJAHR" from metadata_baum.baumkataster where "PFLANZJAHR" > $1;
`

type TreeRegisterRepo struct {
	db *sqlx.DB
}

func NewTreeRegisterRepo(dsn string) (*TreeRegisterRepo, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &TreeRegisterRepo{
		db: db,
	}, nil
}

func (r *TreeRegisterRepo) GetTreesPlantedAfter(ctx context.Context) ([]TreeRegister, error) {
	var trees []TreeRegister
	if err := r.db.SelectContext(ctx, &trees, query, strconv.Itoa(time.Now().Year()-3)); err != nil {
		return nil, err
	}

	return trees, nil
}
