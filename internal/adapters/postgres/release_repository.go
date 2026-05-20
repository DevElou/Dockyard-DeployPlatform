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
		       COALESCE(image_repository, ''), COALESCE(image_tag, ''), COALESCE(image_digest, ''),
		       build_status, created_by_user_id::TEXT, created_at
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

func (r *ReleaseRepository) ListByBuildStatus(ctx context.Context, status domain.BuildStatus) ([]domain.Release, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, version, source_type, git_sha, git_ref,
		       COALESCE(image_repository, ''), COALESCE(image_tag, ''), COALESCE(image_digest, ''),
		       build_status, created_by_user_id::TEXT, created_at
		FROM releases
		WHERE build_status = $1
		ORDER BY created_at ASC
	`, string(status))
	if err != nil {
		return nil, fmt.Errorf("postgres: list releases by build status: %w", err)
	}
	defer rows.Close()

	releases, err := pgx.CollectRows(rows, scanRelease)
	if err != nil {
		return nil, fmt.Errorf("postgres: list releases by build status: %w", err)
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
		VALUES ($1::UUID, $2, $3, $4, $5,
		        NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''),
		        $9, $10::UUID)
		RETURNING id::TEXT, created_at
	`,
		release.ProjectID, release.Version, release.SourceType, release.GitSHA, release.GitRef,
		release.ImageRepository, release.ImageTag, release.ImageDigest,
		string(release.BuildStatus),
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
		       COALESCE(image_repository, ''), COALESCE(image_tag, ''), COALESCE(image_digest, ''),
		       build_status, created_by_user_id::TEXT, created_at
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

func (r *ReleaseRepository) UpdateBuildStatus(ctx context.Context, id string, status domain.BuildStatus) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE releases SET build_status = $1 WHERE id = $2::UUID
	`, string(status), id)
	if err != nil {
		return fmt.Errorf("postgres: update release build status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrReleaseNotFound
	}
	return nil
}

func (r *ReleaseRepository) UpdateBuildResult(ctx context.Context, id, imageRepository, imageTag, imageDigest string, status domain.BuildStatus) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE releases
		SET image_repository = $1, image_tag = $2, image_digest = $3, build_status = $4
		WHERE id = $5::UUID
	`, imageRepository, imageTag, imageDigest, string(status), id)
	if err != nil {
		return fmt.Errorf("postgres: update release build result: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrReleaseNotFound
	}
	return nil
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
