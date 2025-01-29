package main

import (
	"context"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const query = `
select b."OBJECTID", b."BAUMNUMMER", b."HOCHWERT", b."RECHTSWERT", b."GATTUNG", b."GEBIET", b."STRASSE", c."PFLANZJAHR" from (
  select "OBJECTID", cast("PFLANZJAHR" as int) from metadata_baum.baumkataster where "PFLANZJAHR" != 'Null'
) as c 
inner join metadata_baum.baumkataster as b on c."OBJECTID" = b."OBJECTID"
where c."PFLANZJAHR" > $1;
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
