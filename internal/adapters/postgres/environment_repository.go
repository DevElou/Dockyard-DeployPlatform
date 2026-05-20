package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnvironmentSetRepository manages environment sets for projects.
type EnvironmentSetRepository struct {
	pool *pgxpool.Pool
}

func NewEnvironmentSetRepository(pool *pgxpool.Pool) *EnvironmentSetRepository {
	return &EnvironmentSetRepository{pool: pool}
}

func (r *EnvironmentSetRepository) List(ctx context.Context, projectID string) ([]domain.EnvironmentSet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, name
		FROM environment_sets
		WHERE project_id = $1::UUID
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list environment sets: %w", err)
	}
	defer rows.Close()

	sets, err := pgx.CollectRows(rows, scanEnvironmentSet)
	if err != nil {
		return nil, fmt.Errorf("postgres: list environment sets: %w", err)
	}
	if sets == nil {
		return []domain.EnvironmentSet{}, nil
	}
	return sets, nil
}

func (r *EnvironmentSetRepository) Create(ctx context.Context, set domain.EnvironmentSet) (domain.EnvironmentSet, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO environment_sets (project_id, name)
		VALUES ($1::UUID, $2)
		RETURNING id::TEXT
	`, set.ProjectID, set.Name).Scan(&set.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.EnvironmentSet{}, ErrEnvironmentSetNameExists
		}
		return domain.EnvironmentSet{}, fmt.Errorf("postgres: create environment set: %w", err)
	}
	return set, nil
}

func (r *EnvironmentSetRepository) GetByID(ctx context.Context, id string) (domain.EnvironmentSet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, name
		FROM environment_sets
		WHERE id = $1::UUID
	`, id)
	if err != nil {
		return domain.EnvironmentSet{}, fmt.Errorf("postgres: get environment set: %w", err)
	}
	defer rows.Close()

	set, err := pgx.CollectOneRow(rows, scanEnvironmentSet)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.EnvironmentSet{}, ErrEnvironmentSetNotFound
		}
		return domain.EnvironmentSet{}, fmt.Errorf("postgres: get environment set: %w", err)
	}
	return set, nil
}

func (r *EnvironmentSetRepository) GetDefaultForProject(ctx context.Context, projectID string) (domain.EnvironmentSet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, name
		FROM environment_sets
		WHERE project_id = $1::UUID
		ORDER BY created_at ASC
		LIMIT 1
	`, projectID)
	if err != nil {
		return domain.EnvironmentSet{}, fmt.Errorf("postgres: get default environment set: %w", err)
	}
	defer rows.Close()

	set, err := pgx.CollectOneRow(rows, scanEnvironmentSet)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.EnvironmentSet{}, ErrEnvironmentSetNotFound
		}
		return domain.EnvironmentSet{}, fmt.Errorf("postgres: get default environment set: %w", err)
	}
	return set, nil
}

func scanEnvironmentSet(row pgx.CollectableRow) (domain.EnvironmentSet, error) {
	var s domain.EnvironmentSet
	return s, row.Scan(&s.ID, &s.ProjectID, &s.Name)
}

// EnvironmentVariableRepository manages variables within an environment set.
// Values are stored as raw bytes (no encryption for V1 homelab use).
type EnvironmentVariableRepository struct {
	pool *pgxpool.Pool
}

func NewEnvironmentVariableRepository(pool *pgxpool.Pool) *EnvironmentVariableRepository {
	return &EnvironmentVariableRepository{pool: pool}
}

func (r *EnvironmentVariableRepository) ListBySet(ctx context.Context, environmentSetID string) ([]domain.EnvironmentVariable, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, environment_set_id::TEXT, key, value_encrypted, is_secret
		FROM environment_variables
		WHERE environment_set_id = $1::UUID
		ORDER BY key
	`, environmentSetID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list environment variables: %w", err)
	}
	defer rows.Close()

	vars, err := pgx.CollectRows(rows, scanEnvironmentVariable)
	if err != nil {
		return nil, fmt.Errorf("postgres: list environment variables: %w", err)
	}
	if vars == nil {
		return []domain.EnvironmentVariable{}, nil
	}
	return vars, nil
}

func (r *EnvironmentVariableRepository) Upsert(ctx context.Context, environmentSetID, key, value string, isSecret bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO environment_variables (environment_set_id, key, value_encrypted, is_secret)
		VALUES ($1::UUID, $2, $3, $4)
		ON CONFLICT (environment_set_id, key) DO UPDATE
		SET value_encrypted = $3, is_secret = $4, updated_at = now()
	`, environmentSetID, key, []byte(value), isSecret)

	if err != nil {
		return fmt.Errorf("postgres: upsert environment variable: %w", err)
	}
	return nil
}

func (r *EnvironmentVariableRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM environment_variables WHERE id = $1::UUID`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete environment variable: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrEnvironmentVariableNotFound
	}
	return nil
}

func scanEnvironmentVariable(row pgx.CollectableRow) (domain.EnvironmentVariable, error) {
	var v domain.EnvironmentVariable
	var raw []byte
	err := row.Scan(&v.ID, &v.EnvironmentSetID, &v.Key, &raw, &v.IsSecret)
	v.Value = string(raw)
	return v, err
}
