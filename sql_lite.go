package main

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type LocalStorageRepo struct {
	db *sqlx.DB
}

func NewLocalStorageRepo(dsn string) (*LocalStorageRepo, error) {
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	return &LocalStorageRepo{
		db: db,
	}, nil
}

const insertQuery = `
INSERT INTO trees (register_id, planting_year, species, tree_number, latitude, longitude, description) VALUES (:register_id, :planting_year, :species, :tree_number, :latitude, :longitude, :description)
`

func (l *LocalStorageRepo) Insert(ctx context.Context, tree []Tree) error {
	_, err := l.db.NamedExecContext(ctx, insertQuery, tree)
	return err
}

const getAllQuery = `SELECT * FROM trees`

func (l *LocalStorageRepo) GetAll(ctx context.Context) ([]Tree, error) {
	var trees []Tree
	if err := l.db.SelectContext(ctx, &trees, getAllQuery); err != nil {
		return nil, err
	}

	return trees, nil
}

const getAllInQuery = "SELECT * FROM trees WHERE register_id IN (?)"

func (l *LocalStorageRepo) GetAllIn(ctx context.Context, register_ids []string) ([]Tree, error) {
	query, args, err := sqlx.In(getAllInQuery, register_ids)
	if err != nil {
		return nil, err
	}

	query = l.db.Rebind(query)

	var trees []Tree
	if err := l.db.SelectContext(ctx, &trees, query, args...); err != nil {
		return nil, err
	}

	return trees, nil
}

const getAllNotInQuery = "SELECT * FROM trees WHERE register_id NOT IN (?)"

func (l *LocalStorageRepo) GetAllNotIn(ctx context.Context, register_ids []string) ([]Tree, error) {
	query, args, err := sqlx.In(getAllInQuery, register_ids)
	if err != nil {
		return nil, err
	}

	query = l.db.Rebind(query)

	var trees []Tree
	if err := l.db.SelectContext(ctx, &trees, query, args...); err != nil {
		return nil, err
	}

	return trees, nil
}

const updateQuery = `
UPDATE trees SET green_ecolution_id = $2 WHERE register_id = $1;
`

func (l *LocalStorageRepo) Update(ctx context.Context, register_id string, ge_id int32) error {
	_, err := l.db.ExecContext(ctx, updateQuery, register_id, ge_id)
	return err
}

const migration = `
CREATE TABLE IF NOT EXISTS trees (
	register_id TEXT PRIMARY KEY, 
	green_ecolution_id INTEGER UNIQUE, 
	planting_year INTEGER NOT NULL,
	species TEXT NOT NULL,
	tree_number TEXT NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL, 
	description TEXT NOT NULL
)
`

func (l *LocalStorageRepo) Setup(ctx context.Context) error {
	fmt.Print("setup sqlite")
	_, err := l.db.ExecContext(ctx, migration)
	return err
}
