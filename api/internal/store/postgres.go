package store

import (
	"context"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct { DB *pgxpool.Pool }

func New(ctx context.Context, url string) (*Store, error) {
	config, err := pgxpool.ParseConfig(url); if err != nil { return nil, err }
	config.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, config); if err != nil { return nil, err }
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second); defer cancel()
	if err := pool.Ping(pingCtx); err != nil { pool.Close(); return nil, err }
	return &Store{DB: pool}, nil
}

func (s *Store) Close() { s.DB.Close() }
