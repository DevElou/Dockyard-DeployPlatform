package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RuntimeTargetRepository struct {
	pool *pgxpool.Pool
}

func NewRuntimeTargetRepository(pool *pgxpool.Pool) *RuntimeTargetRepository {
	return &RuntimeTargetRepository{pool: pool}
}

func (r *RuntimeTargetRepository) List(ctx context.Context) ([]domain.RuntimeTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, slug, name, runtime_type, endpoint, agent_key_hash,
		       server_group, region, enabled
		FROM runtime_targets
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("postgres: list runtime targets: %w", err)
	}
	defer rows.Close()

	targets, err := pgx.CollectRows(rows, scanRuntimeTarget)
	if err != nil {
		return nil, fmt.Errorf("postgres: list runtime targets: %w", err)
	}
	if targets == nil {
		return []domain.RuntimeTarget{}, nil
	}
	return targets, nil
}

func (r *RuntimeTargetRepository) Create(ctx context.Context, rt domain.RuntimeTarget) (domain.RuntimeTarget, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO runtime_targets (slug, name, runtime_type, endpoint, agent_key_hash, server_group, region, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::TEXT
	`,
		rt.Slug, rt.Name, string(rt.RuntimeType), rt.Endpoint, rt.AgentKeyHash,
		nullableString(rt.ServerGroup), nullableString(rt.Region), rt.Enabled,
	).Scan(&rt.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.RuntimeTarget{}, ErrRuntimeTargetSlugExists
		}
		return domain.RuntimeTarget{}, fmt.Errorf("postgres: create runtime target: %w", err)
	}

	return rt, nil
}

func (r *RuntimeTargetRepository) GetByID(ctx context.Context, id string) (domain.RuntimeTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, slug, name, runtime_type, endpoint, agent_key_hash,
		       server_group, region, enabled
		FROM runtime_targets
		WHERE id = $1::UUID
	`, id)
	if err != nil {
		return domain.RuntimeTarget{}, fmt.Errorf("postgres: get runtime target: %w", err)
	}
	defer rows.Close()

	rt, err := pgx.CollectOneRow(rows, scanRuntimeTarget)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RuntimeTarget{}, ErrRuntimeTargetNotFound
		}
		return domain.RuntimeTarget{}, fmt.Errorf("postgres: get runtime target: %w", err)
	}
	return rt, nil
}

func (r *RuntimeTargetRepository) SetEnabled(ctx context.Context, id string, enabled bool) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE runtime_targets SET enabled = $1, updated_at = now() WHERE id = $2::UUID
	`, enabled, id)
	if err != nil {
		return fmt.Errorf("postgres: set runtime target enabled: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRuntimeTargetNotFound
	}
	return nil
}

func scanRuntimeTarget(row pgx.CollectableRow) (domain.RuntimeTarget, error) {
	var rt domain.RuntimeTarget
	var runtimeType string
	var serverGroup pgtype.Text
	var region pgtype.Text

	err := row.Scan(
		&rt.ID, &rt.Slug, &rt.Name, &runtimeType, &rt.Endpoint, &rt.AgentKeyHash,
		&serverGroup, &region, &rt.Enabled,
	)
	if err != nil {
		return domain.RuntimeTarget{}, err
	}

	rt.RuntimeType = domain.RuntimeType(runtimeType)
	if serverGroup.Valid {
		v := serverGroup.String
		rt.ServerGroup = &v
	}
	if region.Valid {
		v := region.String
		rt.Region = &v
	}
	return rt, nil
}

func nullableString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}
