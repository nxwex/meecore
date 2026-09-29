package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nxwex/meecore/internal/node"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) Get(ctx context.Context, id int64) (*node.Node, error) {
	raw := `SELECT id, name, hostname, ip_address, created_at
			FROM nodes
			WHERE id = $1`

	n := &node.Node{}

	if err := s.db.QueryRow(ctx, raw, id).Scan(
		&n.ID,
		&n.Name,
		&n.Hostname,
		&n.IPAddress,
		&n.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, node.ErrNotFound
		}

		return nil, fmt.Errorf("get node %d: %w", id, err)
	}

	return n, nil
}
