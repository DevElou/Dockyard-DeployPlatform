package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeploymentRepository struct {
	pool *pgxpool.Pool
}

func NewDeploymentRepository(pool *pgxpool.Pool) *DeploymentRepository {
	return &DeploymentRepository{pool: pool}
}

func (r *DeploymentRepository) List(ctx context.Context, projectID string) ([]domain.Deployment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, release_id::TEXT, runtime_target_id::TEXT,
		       project_service_id::TEXT, environment_set_id::TEXT,
		       status, strategy, triggered_by_user_id::TEXT,
		       rollback_of_deployment_id::TEXT, started_at, finished_at, created_at
		FROM deployments
		WHERE project_id = $1::UUID
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list deployments: %w", err)
	}
	defer rows.Close()

	deployments, err := pgx.CollectRows(rows, scanDeployment)
	if err != nil {
		return nil, fmt.Errorf("postgres: list deployments: %w", err)
	}
	if deployments == nil {
		return []domain.Deployment{}, nil
	}
	return deployments, nil
}

func (r *DeploymentRepository) Create(ctx context.Context, d domain.Deployment) (domain.Deployment, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO deployments (project_id, release_id, runtime_target_id,
		                         project_service_id, environment_set_id,
		                         status, strategy, triggered_by_user_id,
		                         rollback_of_deployment_id)
		VALUES ($1::UUID, $2::UUID, $3::UUID, $4::UUID, $5::UUID, $6, $7, $8::UUID, $9::UUID)
		RETURNING id::TEXT, created_at
	`,
		d.ProjectID, d.ReleaseID, d.RuntimeTargetID,
		nullableString(d.ProjectServiceID), nullableString(d.EnvironmentSetID),
		string(d.Status), d.Strategy,
		nullableString(d.TriggeredByUserID), nullableString(d.RollbackOfDeploymentID),
	).Scan(&d.ID, &d.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "deployments_release_id_fkey":
				return domain.Deployment{}, ErrReleaseNotFound
			case "deployments_runtime_target_id_fkey":
				return domain.Deployment{}, ErrRuntimeTargetNotFound
			default:
				return domain.Deployment{}, ErrProjectNotFound
			}
		}
		return domain.Deployment{}, fmt.Errorf("postgres: create deployment: %w", err)
	}

	return d, nil
}

func (r *DeploymentRepository) GetByID(ctx context.Context, id string) (domain.Deployment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, release_id::TEXT, runtime_target_id::TEXT,
		       project_service_id::TEXT, environment_set_id::TEXT,
		       status, strategy, triggered_by_user_id::TEXT,
		       rollback_of_deployment_id::TEXT, started_at, finished_at, created_at
		FROM deployments
		WHERE id = $1::UUID
	`, id)
	if err != nil {
		return domain.Deployment{}, fmt.Errorf("postgres: get deployment: %w", err)
	}
	defer rows.Close()

	d, err := pgx.CollectOneRow(rows, scanDeployment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Deployment{}, ErrDeploymentNotFound
		}
		return domain.Deployment{}, fmt.Errorf("postgres: get deployment: %w", err)
	}
	return d, nil
}

func (r *DeploymentRepository) UpdateStatus(ctx context.Context, id string, status domain.DeploymentStatus, startedAt *time.Time, finishedAt *time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE deployments
		SET status = $1,
		    started_at = COALESCE($2, started_at),
		    finished_at = COALESCE($3, finished_at)
		WHERE id = $4::UUID
	`, string(status), nullableTime(startedAt), nullableTime(finishedAt), id)
	if err != nil {
		return fmt.Errorf("postgres: update deployment status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDeploymentNotFound
	}
	return nil
}

func scanDeployment(row pgx.CollectableRow) (domain.Deployment, error) {
	var d domain.Deployment
	var status string
	var projectServiceID, environmentSetID, triggeredByUserID, rollbackOfDeploymentID pgtype.Text
	var startedAt, finishedAt pgtype.Timestamptz

	err := row.Scan(
		&d.ID, &d.ProjectID, &d.ReleaseID, &d.RuntimeTargetID,
		&projectServiceID, &environmentSetID,
		&status, &d.Strategy, &triggeredByUserID,
		&rollbackOfDeploymentID, &startedAt, &finishedAt, &d.CreatedAt,
	)
	if err != nil {
		return domain.Deployment{}, err
	}

	d.Status = domain.DeploymentStatus(status)
	if projectServiceID.Valid {
		v := projectServiceID.String
		d.ProjectServiceID = &v
	}
	if environmentSetID.Valid {
		v := environmentSetID.String
		d.EnvironmentSetID = &v
	}
	if triggeredByUserID.Valid {
		v := triggeredByUserID.String
		d.TriggeredByUserID = &v
	}
	if rollbackOfDeploymentID.Valid {
		v := rollbackOfDeploymentID.String
		d.RollbackOfDeploymentID = &v
	}
	if startedAt.Valid {
		v := startedAt.Time
		d.StartedAt = &v
	}
	if finishedAt.Valid {
		v := finishedAt.Time
		d.FinishedAt = &v
	}
	return d, nil
}

func nullableTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
