package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/http/handlers/redirect"
	"url-shortener/internal/http/handlers/save"
	"url-shortener/internal/storage/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DSN базы данных из нашего docker-compose
const storagePath = "postgres://user:password@localhost:5432/shortener?sslmode=disable"

func main() {
	// 1. Настраиваем красивый логгер
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log.Info("starting url-shortener")

	// 2. Накатываем миграции (создаем таблицы)
	m, err := migrate.New("file://migrations", storagePath)
	if err != nil {
		log.Error("failed to init migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("no migrations to apply")
		} else {
			log.Error("failed to apply migrations", slog.String("error", err.Error()))
			os.Exit(1)
		}
	} else {
		log.Info("migrations applied successfully")
	}

	// 3. Подключаемся к базе (пул соединений)
	pool, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		log.Error("failed to connect to db", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Пингуем базу, чтобы убедиться, что она реально жива
	if err := pool.Ping(context.Background()); err != nil {
		log.Error("failed to ping db", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("DB connection established successfully! Monster is awake.")

	store := postgres.New(pool)

	log.Info("starting server", slog.String("port", "8080"))

	mux := http.NewServeMux()

	// Передаем наш store в хендлер.яё действии!
	mux.HandleFunc("POST /save", save.New(log, store))

	mux.HandleFunc("GET /{alias}", redirect.New(log, store))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", slog.String("error", err.Error()))
	}

	// 4. Проверяем ошибку NotFound (алиас, которого нет)
	_, err = store.GetURL(context.Background(), "not_exist")
	if err != nil {
		log.Info("expected error for missing url", slog.String("error", err.Error()))
	} else {
		log.Error("WAIT, WHY NO ERROR FOR MISSING URL?!")
	}

}
