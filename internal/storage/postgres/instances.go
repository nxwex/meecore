package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nxwex/meecore/internal/server"
)

func (s *Storage) CreateInstance(ctx context.Context, instance *server.Instance) error {
	raw := `
		INSERT INTO servers (
			node_id,
			name,
			template,
			container_id
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	if err := s.db.QueryRow(
		ctx,
		raw,
		instance.NodeID,
		instance.Name,
		instance.Template,
		instance.ContainerID,
	).Scan(
		&instance.ID,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create instance: %w", err)
	}

	return nil
}

func (s *Storage) GetInstance(ctx context.Context, id int64) (*server.Instance, error) {
	raw := `
		SELECT
			id,
			node_id,
			name,
			template,
			container_id,
			created_at,
			updated_at,
			is_deleted
		FROM servers
		WHERE id = $1
		  AND is_deleted = FALSE
	`

	instance := &server.Instance{}

	if err := s.db.QueryRow(ctx, raw, id).Scan(
		&instance.ID,
		&instance.NodeID,
		&instance.Name,
		&instance.Template,
		&instance.ContainerID,
		&instance.CreatedAt,
		&instance.UpdatedAt,
		&instance.IsDeleted,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, server.ErrInstanceNotFound
		}

		return nil, fmt.Errorf("get instance %s: %w", id, err)
	}

	return instance, nil
}

func (s *Storage) ListInstances(ctx context.Context) ([]server.Instance, error) {
	raw := `
		SELECT
			id,
			node_id,
			name,
			template,
			container_id,
			created_at,
			updated_at,
			is_deleted
		FROM servers
		WHERE is_deleted = FALSE
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	defer rows.Close()

	instances := make([]server.Instance, 0)

	for rows.Next() {
		var instance server.Instance

		if err := rows.Scan(
			&instance.ID,
			&instance.NodeID,
			&instance.Name,
			&instance.Template,
			&instance.ContainerID,
			&instance.CreatedAt,
			&instance.UpdatedAt,
			&instance.IsDeleted,
		); err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}

		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate instances: %w", err)
	}

	return instances, nil
}

func (s *Storage) DeleteInstance(ctx context.Context, id int64) error {
	raw := `
		UPDATE servers
		SET
			is_deleted = TRUE,
			updated_at = NOW()
		WHERE id = $1
		  AND is_deleted = FALSE
	`

	result, err := s.db.Exec(ctx, raw, id)
	if err != nil {
		return fmt.Errorf("delete instance %s: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return server.ErrInstanceNotFound
	}

	return nil
}
