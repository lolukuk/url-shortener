package postgres

import (
	"context"
	"errors"
	"fmt"

	"url-shortener/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

// SaveURL сохраняет длинную ссылку и алиас в базу.
func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	// ВПИШИ SQL ЗАПРОС №1: INSERT INTO ... RETURNING id
	query := `INSERT INTO urls (original_url, alias) VALUES ($1, $2) RETURNING id`

	var id int64
	err := s.db.QueryRow(ctx, query, urlToSave, alias).Scan(&id)
	if err != nil {
		// Проверяем, не является ли ошибка нарушением уникальности (алиас уже занят)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// GetURL достает длинную ссылку по её алиасу.
func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	// ВПИШИ SQL ЗАПРОС №2: SELECT original_url FROM ...
	query := `SELECT original_url FROM urls WHERE alias = $1`

	var resURL string
	err := s.db.QueryRow(ctx, query, alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}
