CREATE TABLE operation_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  resource_type STRING NOT NULL,
  resource_id UUID NOT NULL,
  phase STRING NOT NULL,
  level STRING NOT NULL,
  message STRING NOT NULL,
  details JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_operation_events_resource_created_at
  ON operation_events (resource_type, resource_id, created_at DESC);
