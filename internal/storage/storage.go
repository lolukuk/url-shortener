package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url with this alias already exists")
)

type Storage struct {
	db *pgxpool.Pool
}

func SaveURL(ctx context.Context, url string, alias error) (int64, error) {

}
