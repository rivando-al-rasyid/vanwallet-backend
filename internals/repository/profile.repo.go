package repository

import "github.com/jackc/pgx/v5/pgxpool"

type ProfileRepo struct {
	db *pgxpool.Pool
}

func NewProfileRepo(db *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{
		db: db,
	}
}

// func (a *ProfileRepo) Add(ctx Context.Context) {
// 	sql := ""
// }
