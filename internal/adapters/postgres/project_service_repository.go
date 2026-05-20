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

type ProjectServiceRepository struct {
	pool *pgxpool.Pool
}

func NewProjectServiceRepository(pool *pgxpool.Pool) *ProjectServiceRepository {
	return &ProjectServiceRepository{pool: pool}
}

func (r *ProjectServiceRepository) List(ctx context.Context, projectID string) ([]domain.ProjectService, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, name, container_port,
		       COALESCE(healthcheck_path, ''), COALESCE(healthcheck_port, 0), routing_enabled
		FROM project_services
		WHERE project_id = $1::UUID
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list project services: %w", err)
	}
	defer rows.Close()

	services, err := pgx.CollectRows(rows, scanProjectService)
	if err != nil {
		return nil, fmt.Errorf("postgres: list project services: %w", err)
	}
	if services == nil {
		return []domain.ProjectService{}, nil
	}
	return services, nil
}

func (r *ProjectServiceRepository) Create(ctx context.Context, ps domain.ProjectService) (domain.ProjectService, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO project_services (project_id, name, container_port, healthcheck_path, healthcheck_port, routing_enabled)
		VALUES ($1::UUID, $2, $3, NULLIF($4, ''), NULLIF($5, 0), $6)
		RETURNING id::TEXT, project_id::TEXT
	`,
		ps.ProjectID, ps.Name, ps.ContainerPort,
		ps.HealthcheckPath, ps.HealthcheckPort, ps.RoutingEnabled,
	).Scan(&ps.ID, &ps.ProjectID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ProjectService{}, ErrProjectServiceNameExists
		}
		return domain.ProjectService{}, fmt.Errorf("postgres: create project service: %w", err)
	}
	return ps, nil
}

func (r *ProjectServiceRepository) GetByID(ctx context.Context, id string) (domain.ProjectService, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, name, container_port,
		       COALESCE(healthcheck_path, ''), COALESCE(healthcheck_port, 0), routing_enabled
		FROM project_services
		WHERE id = $1::UUID
	`, id)
	if err != nil {
		return domain.ProjectService{}, fmt.Errorf("postgres: get project service: %w", err)
	}
	defer rows.Close()

	ps, err := pgx.CollectOneRow(rows, scanProjectService)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectService{}, ErrProjectServiceNotFound
		}
		return domain.ProjectService{}, fmt.Errorf("postgres: get project service: %w", err)
	}
	return ps, nil
}

func (r *ProjectServiceRepository) GetDefaultForProject(ctx context.Context, projectID string) (domain.ProjectService, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, name, container_port,
		       COALESCE(healthcheck_path, ''), COALESCE(healthcheck_port, 0), routing_enabled
		FROM project_services
		WHERE project_id = $1::UUID
		ORDER BY created_at ASC
		LIMIT 1
	`, projectID)
	if err != nil {
		return domain.ProjectService{}, fmt.Errorf("postgres: get default project service: %w", err)
	}
	defer rows.Close()

	ps, err := pgx.CollectOneRow(rows, scanProjectService)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectService{}, ErrProjectServiceNotFound
		}
		return domain.ProjectService{}, fmt.Errorf("postgres: get default project service: %w", err)
	}
	return ps, nil
}

func scanProjectService(row pgx.CollectableRow) (domain.ProjectService, error) {
	var ps domain.ProjectService
	return ps, row.Scan(
		&ps.ID, &ps.ProjectID, &ps.Name, &ps.ContainerPort,
		&ps.HealthcheckPath, &ps.HealthcheckPort, &ps.RoutingEnabled,
	)
}
