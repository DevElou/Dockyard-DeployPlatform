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

type ReleaseRepository struct {
	pool *pgxpool.Pool
}

func NewReleaseRepository(pool *pgxpool.Pool) *ReleaseRepository {
	return &ReleaseRepository{pool: pool}
}

func (r *ReleaseRepository) List(ctx context.Context, projectID string) ([]domain.Release, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, version, source_type, git_sha, git_ref,
		       image_repository, image_tag, image_digest, build_status,
		       created_by_user_id::TEXT, created_at
		FROM releases
		WHERE project_id = $1::UUID
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list releases: %w", err)
	}
	defer rows.Close()

	releases, err := pgx.CollectRows(rows, scanRelease)
	if err != nil {
		return nil, fmt.Errorf("postgres: list releases: %w", err)
	}
	if releases == nil {
		return []domain.Release{}, nil
	}
	return releases, nil
}

func (r *ReleaseRepository) Create(ctx context.Context, release domain.Release) (domain.Release, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO releases (project_id, version, source_type, git_sha, git_ref,
		                      image_repository, image_tag, image_digest, build_status, created_by_user_id)
		VALUES ($1::UUID, $2, $3, $4, $5, $6, $7, $8, $9, $10::UUID)
		RETURNING id::TEXT, created_at
	`,
		release.ProjectID, release.Version, release.SourceType, release.GitSHA, release.GitRef,
		release.ImageRepository, release.ImageTag, release.ImageDigest, string(release.BuildStatus),
		nullableString(release.CreatedByUserID),
	).Scan(&release.ID, &release.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "releases_project_id_image_digest_key":
				return domain.Release{}, ErrReleaseDigestExists
			default:
				return domain.Release{}, ErrReleaseVersionExists
			}
		}
		return domain.Release{}, fmt.Errorf("postgres: create release: %w", err)
	}

	return release, nil
}

func (r *ReleaseRepository) GetByID(ctx context.Context, id string) (domain.Release, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, version, source_type, git_sha, git_ref,
		       image_repository, image_tag, image_digest, build_status,
		       created_by_user_id::TEXT, created_at
		FROM releases
		WHERE id = $1::UUID
	`, id)
	if err != nil {
		return domain.Release{}, fmt.Errorf("postgres: get release: %w", err)
	}
	defer rows.Close()

	release, err := pgx.CollectOneRow(rows, scanRelease)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Release{}, ErrReleaseNotFound
		}
		return domain.Release{}, fmt.Errorf("postgres: get release: %w", err)
	}
	return release, nil
}

func scanRelease(row pgx.CollectableRow) (domain.Release, error) {
	var rel domain.Release
	var buildStatus string
	var createdByUserID pgtype.Text

	err := row.Scan(
		&rel.ID, &rel.ProjectID, &rel.Version, &rel.SourceType, &rel.GitSHA, &rel.GitRef,
		&rel.ImageRepository, &rel.ImageTag, &rel.ImageDigest, &buildStatus,
		&createdByUserID, &rel.CreatedAt,
	)
	if err != nil {
		return domain.Release{}, err
	}

	rel.BuildStatus = domain.BuildStatus(buildStatus)
	if createdByUserID.Valid {
		v := createdByUserID.String
		rel.CreatedByUserID = &v
	}
	return rel, nil
}
