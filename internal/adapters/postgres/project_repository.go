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

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, slug, name, status, github_owner, github_repo,
		       default_branch, root_directory, dockerfile_path, build_context
		FROM projects
		WHERE status != 'archived'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("postgres: list projects: %w", err)
	}
	defer rows.Close()

	projects, err := pgx.CollectRows(rows, scanProject)
	if err != nil {
		return nil, fmt.Errorf("postgres: list projects: %w", err)
	}
	if projects == nil {
		return []domain.Project{}, nil
	}
	return projects, nil
}

func (r *ProjectRepository) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO projects (slug, name, github_owner, github_repo, default_branch, root_directory, dockerfile_path, build_context)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::TEXT
	`,
		project.Slug, project.Name, project.GitHubOwner, project.GitHubRepo,
		project.DefaultBranch, project.RootDirectory, project.DockerfilePath, project.BuildContext,
	).Scan(&project.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Project{}, ErrProjectSlugExists
		}
		return domain.Project{}, fmt.Errorf("postgres: create project: %w", err)
	}

	project.Status = domain.ProjectStatusActive
	return project, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (domain.Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, slug, name, status, github_owner, github_repo,
		       default_branch, root_directory, dockerfile_path, build_context
		FROM projects
		WHERE id = $1::UUID
	`, id)
	if err != nil {
		return domain.Project{}, fmt.Errorf("postgres: get project: %w", err)
	}
	defer rows.Close()

	p, err := pgx.CollectOneRow(rows, scanProject)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrProjectNotFound
		}
		return domain.Project{}, fmt.Errorf("postgres: get project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, slug, name, status, github_owner, github_repo,
		       default_branch, root_directory, dockerfile_path, build_context
		FROM projects
		WHERE slug = $1
	`, slug)
	if err != nil {
		return domain.Project{}, fmt.Errorf("postgres: get project by slug: %w", err)
	}
	defer rows.Close()

	p, err := pgx.CollectOneRow(rows, scanProject)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrProjectNotFound
		}
		return domain.Project{}, fmt.Errorf("postgres: get project by slug: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) Archive(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE projects SET status = 'archived', updated_at = now() WHERE id = $1::UUID
	`, id)
	if err != nil {
		return fmt.Errorf("postgres: archive project: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) ListRuntimeTargets(ctx context.Context, projectID string) ([]domain.RuntimeTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rt.id::TEXT, rt.slug, rt.name, rt.runtime_type, rt.endpoint, rt.agent_key_hash,
		       rt.server_group, rt.region, rt.enabled
		FROM runtime_targets rt
		JOIN project_runtime_targets prt ON prt.runtime_target_id = rt.id
		WHERE prt.project_id = $1::UUID
		ORDER BY prt.created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list project runtime targets: %w", err)
	}
	defer rows.Close()

	targets, err := pgx.CollectRows(rows, scanRuntimeTarget)
	if err != nil {
		return nil, fmt.Errorf("postgres: list project runtime targets: %w", err)
	}
	if targets == nil {
		return []domain.RuntimeTarget{}, nil
	}
	return targets, nil
}

func (r *ProjectRepository) AddRuntimeTarget(ctx context.Context, projectID, runtimeTargetID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO project_runtime_targets (project_id, runtime_target_id)
		VALUES ($1::UUID, $2::UUID)
	`, projectID, runtimeTargetID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return ErrProjectRuntimeTargetExists
			case "23503":
				if pgErr.ConstraintName == "project_runtime_targets_project_id_fkey" {
					return ErrProjectNotFound
				}
				return ErrRuntimeTargetNotFound
			}
		}
		return fmt.Errorf("postgres: add runtime target to project: %w", err)
	}
	return nil
}

func scanProject(row pgx.CollectableRow) (domain.Project, error) {
	var p domain.Project
	var status string
	return p, row.Scan(
		&p.ID, &p.Slug, &p.Name, &status, &p.GitHubOwner, &p.GitHubRepo,
		&p.DefaultBranch, &p.RootDirectory, &p.DockerfilePath, &p.BuildContext,
	)
}
