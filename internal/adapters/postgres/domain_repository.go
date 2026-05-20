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

type DomainRepository struct {
	pool *pgxpool.Pool
}

func NewDomainRepository(pool *pgxpool.Pool) *DomainRepository {
	return &DomainRepository{pool: pool}
}

func (r *DomainRepository) List(ctx context.Context, projectID string) ([]domain.Domain, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, project_service_id::TEXT,
		       hostname, base_domain, provider, routing_type, tls_enabled, status
		FROM domains
		WHERE project_id = $1::UUID
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list domains: %w", err)
	}
	defer rows.Close()

	domains, err := pgx.CollectRows(rows, scanDomain)
	if err != nil {
		return nil, fmt.Errorf("postgres: list domains: %w", err)
	}
	if domains == nil {
		return []domain.Domain{}, nil
	}
	return domains, nil
}

func (r *DomainRepository) Create(ctx context.Context, d domain.Domain) (domain.Domain, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO domains (project_id, project_service_id, hostname, base_domain,
		                     provider, routing_type, tls_enabled, status)
		VALUES ($1::UUID, $2::UUID, $3, $4, $5, $6, $7, $8)
		RETURNING id::TEXT
	`,
		d.ProjectID, nullableString(d.ProjectServiceID),
		d.Hostname, d.BaseDomain, d.Provider, d.RoutingType, d.TLSEnabled, string(d.Status),
	).Scan(&d.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Domain{}, ErrDomainHostnameExists
		}
		return domain.Domain{}, fmt.Errorf("postgres: create domain: %w", err)
	}

	return d, nil
}

func (r *DomainRepository) GetByID(ctx context.Context, id string) (domain.Domain, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, project_id::TEXT, project_service_id::TEXT,
		       hostname, base_domain, provider, routing_type, tls_enabled, status
		FROM domains
		WHERE id = $1::UUID
	`, id)
	if err != nil {
		return domain.Domain{}, fmt.Errorf("postgres: get domain: %w", err)
	}
	defer rows.Close()

	d, err := pgx.CollectOneRow(rows, scanDomain)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Domain{}, ErrDomainNotFound
		}
		return domain.Domain{}, fmt.Errorf("postgres: get domain: %w", err)
	}
	return d, nil
}

func (r *DomainRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM domains WHERE id = $1::UUID`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete domain: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDomainNotFound
	}
	return nil
}

func scanDomain(row pgx.CollectableRow) (domain.Domain, error) {
	var d domain.Domain
	var status string
	var projectServiceID pgtype.Text

	err := row.Scan(
		&d.ID, &d.ProjectID, &projectServiceID,
		&d.Hostname, &d.BaseDomain, &d.Provider, &d.RoutingType, &d.TLSEnabled, &status,
	)
	if err != nil {
		return domain.Domain{}, err
	}

	d.Status = domain.DomainStatus(status)
	if projectServiceID.Valid {
		v := projectServiceID.String
		d.ProjectServiceID = &v
	}
	return d, nil
}
