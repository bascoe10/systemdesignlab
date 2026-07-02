// Package store is the PostgreSQL persistence layer. The schema is created
// by system/db/init.sql when the postgres container first starts.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/systemdesignlab/url-shortener/internal/config"
	"github.com/systemdesignlab/url-shortener/internal/obs"
)

// ErrNotFound is returned when a short code has no mapping.
var ErrNotFound = errors.New("store: not found")

// ErrConflict is returned when a short code already exists.
var ErrConflict = errors.New("store: code already exists")

type Store struct {
	db *sql.DB
}

// Open connects with the pool size from config and starts a goroutine that
// exports pool saturation gauges (the Saturation golden signal).
func Open(cfg *config.Config) (*Store, error) {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.Database.PoolSize)
	db.SetMaxIdleConns(cfg.Database.PoolSize)
	obs.DBPoolMax.Set(float64(cfg.Database.PoolSize))

	// Wait for postgres to accept connections (compose healthcheck races).
	deadline := time.Now().Add(60 * time.Second)
	for {
		err = db.Ping()
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("database not reachable: %w", err)
		}
		time.Sleep(time.Second)
	}

	go func() {
		for range time.Tick(5 * time.Second) {
			obs.DBPoolInUse.Set(float64(db.Stats().InUse))
		}
	}()
	return &Store{db: db}, nil
}

func (s *Store) Insert(ctx context.Context, code, target string) error {
	start := time.Now()
	defer func() { obs.DBQueryDuration.WithLabelValues("insert_url").Observe(time.Since(start).Seconds()) }()

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO urls (code, target) VALUES ($1, $2) ON CONFLICT (code) DO NOTHING`, code, target)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConflict
	}
	return nil
}

func (s *Store) Get(ctx context.Context, code string) (string, error) {
	start := time.Now()
	defer func() { obs.DBQueryDuration.WithLabelValues("get_url").Observe(time.Since(start).Seconds()) }()

	var target string
	err := s.db.QueryRowContext(ctx, `SELECT target FROM urls WHERE code = $1`, code).Scan(&target)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return target, err
}

func (s *Store) Healthy(ctx context.Context) bool { return s.db.PingContext(ctx) == nil }

func (s *Store) Close() error { return s.db.Close() }
