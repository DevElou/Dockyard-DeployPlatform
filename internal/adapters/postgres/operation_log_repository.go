package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultOperationEventRetention bounds the persisted history per resource.
const DefaultOperationEventRetention = 1000

type OperationLogRepository struct {
	pool      *pgxpool.Pool
	retention int
}

func NewOperationLogRepository(pool *pgxpool.Pool) *OperationLogRepository {
	return &OperationLogRepository{pool: pool, retention: DefaultOperationEventRetention}
}

// WithRetention overrides the per-resource retention bound (mainly for tests).
func (r *OperationLogRepository) WithRetention(keep int) *OperationLogRepository {
	clone := *r
	clone.retention = keep
	return &clone
}

func (r *OperationLogRepository) Append(ctx context.Context, event domain.OperationEvent) (domain.OperationEvent, error) {
	if err := event.Validate(); err != nil {
		return domain.OperationEvent{}, fmt.Errorf("postgres: append operation event: %w", err)
	}

	detailsJSON, err := encodeDetails(event.Details)
	if err != nil {
		return domain.OperationEvent{}, fmt.Errorf("postgres: encode operation event details: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO operation_events (resource_type, resource_id, phase, level, message, details)
		VALUES ($1, $2::UUID, $3, $4, $5, $6::JSONB)
		RETURNING id::TEXT, created_at
	`,
		string(event.ResourceType), event.ResourceID, event.Phase,
		string(event.Level), event.Message, detailsJSON,
	).Scan(&event.ID, &event.CreatedAt)
	if err != nil {
		return domain.OperationEvent{}, fmt.Errorf("postgres: append operation event: %w", err)
	}

	if r.retention > 0 {
		if pruneErr := r.Prune(ctx, event.ResourceType, event.ResourceID, r.retention); pruneErr != nil {
			return event, fmt.Errorf("postgres: prune operation events after append: %w", pruneErr)
		}
	}
	return event, nil
}

func (r *OperationLogRepository) List(ctx context.Context, resourceType domain.OperationResourceType, resourceID string) ([]domain.OperationEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::TEXT, resource_type, resource_id::TEXT, phase, level, message,
		       COALESCE(details::TEXT, ''), created_at
		FROM operation_events
		WHERE resource_type = $1 AND resource_id = $2::UUID
		ORDER BY created_at ASC, id ASC
	`, string(resourceType), resourceID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list operation events: %w", err)
	}
	defer rows.Close()

	events, err := pgx.CollectRows(rows, scanOperationEvent)
	if err != nil {
		return nil, fmt.Errorf("postgres: list operation events: %w", err)
	}
	if events == nil {
		return []domain.OperationEvent{}, nil
	}
	return events, nil
}

func (r *OperationLogRepository) Prune(ctx context.Context, resourceType domain.OperationResourceType, resourceID string, keep int) error {
	if keep <= 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		DELETE FROM operation_events
		WHERE resource_type = $1 AND resource_id = $2::UUID
		  AND id NOT IN (
		    SELECT id FROM operation_events
		    WHERE resource_type = $1 AND resource_id = $2::UUID
		    ORDER BY created_at DESC, id DESC
		    LIMIT $3
		  )
	`, string(resourceType), resourceID, keep)
	if err != nil {
		return fmt.Errorf("postgres: prune operation events: %w", err)
	}
	return nil
}

func encodeDetails(details map[string]string) ([]byte, error) {
	if len(details) == 0 {
		return nil, nil
	}
	return json.Marshal(details)
}

func decodeDetails(raw string) (map[string]string, error) {
	if raw == "" {
		return nil, nil
	}
	out := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func scanOperationEvent(row pgx.CollectableRow) (domain.OperationEvent, error) {
	var e domain.OperationEvent
	var resourceType, level, detailsRaw string

	if err := row.Scan(
		&e.ID, &resourceType, &e.ResourceID, &e.Phase, &level, &e.Message,
		&detailsRaw, &e.CreatedAt,
	); err != nil {
		return domain.OperationEvent{}, err
	}

	e.ResourceType = domain.OperationResourceType(resourceType)
	e.Level = domain.OperationLevel(level)

	details, err := decodeDetails(detailsRaw)
	if err != nil {
		return domain.OperationEvent{}, fmt.Errorf("postgres: decode operation event details: %w", err)
	}
	e.Details = details
	return e, nil
}
