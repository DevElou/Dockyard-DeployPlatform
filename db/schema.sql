CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email STRING NOT NULL UNIQUE,
  display_name STRING NOT NULL,
  role STRING NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE github_installations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  provider STRING NOT NULL DEFAULT 'github',
  installation_id STRING NOT NULL UNIQUE,
  account_name STRING NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug STRING NOT NULL UNIQUE,
  name STRING NOT NULL,
  description STRING,
  github_installation_id UUID REFERENCES github_installations(id),
  github_owner STRING NOT NULL,
  github_repo STRING NOT NULL,
  default_branch STRING NOT NULL,
  root_directory STRING NOT NULL DEFAULT '.',
  dockerfile_path STRING NOT NULL DEFAULT 'Dockerfile',
  build_context STRING NOT NULL DEFAULT '.',
  status STRING NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE runtime_targets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug STRING NOT NULL UNIQUE,
  name STRING NOT NULL,
  runtime_type STRING NOT NULL DEFAULT 'docker',
  endpoint STRING NOT NULL,
  agent_key_hash STRING NOT NULL,
  server_group STRING,
  region STRING,
  enabled BOOL NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE project_runtime_targets (
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  runtime_target_id UUID NOT NULL REFERENCES runtime_targets(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, runtime_target_id)
);

CREATE TABLE environment_sets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name STRING NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, name)
);

CREATE TABLE environment_variables (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  environment_set_id UUID NOT NULL REFERENCES environment_sets(id) ON DELETE CASCADE,
  key STRING NOT NULL,
  value_encrypted BYTES NOT NULL,
  is_secret BOOL NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (environment_set_id, key)
);

CREATE TABLE project_services (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name STRING NOT NULL,
  container_port INT NOT NULL,
  healthcheck_path STRING,
  healthcheck_port INT,
  routing_enabled BOOL NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, name)
);

CREATE TABLE releases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  version STRING NOT NULL,
  source_type STRING NOT NULL DEFAULT 'github',
  git_sha STRING NOT NULL,
  git_ref STRING NOT NULL,
  image_repository STRING,
  image_tag STRING,
  image_digest STRING,
  build_status STRING NOT NULL,
  created_by_user_id UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, version),
  UNIQUE (project_id, image_digest)
);

CREATE TABLE build_jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  git_sha STRING NOT NULL,
  status STRING NOT NULL,
  logs_ref STRING,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE deployments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  release_id UUID NOT NULL REFERENCES releases(id),
  runtime_target_id UUID NOT NULL REFERENCES runtime_targets(id),
  project_service_id UUID REFERENCES project_services(id),
  environment_set_id UUID REFERENCES environment_sets(id),
  status STRING NOT NULL,
  strategy STRING NOT NULL DEFAULT 'recreate',
  triggered_by_user_id UUID REFERENCES users(id),
  rollback_of_deployment_id UUID REFERENCES deployments(id),
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE deployment_steps (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
  step_type STRING NOT NULL,
  status STRING NOT NULL,
  message STRING,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE domains (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  project_service_id UUID REFERENCES project_services(id),
  hostname STRING NOT NULL UNIQUE,
  base_domain STRING NOT NULL,
  provider STRING NOT NULL DEFAULT 'duckdns',
  routing_type STRING NOT NULL DEFAULT 'host',
  tls_enabled BOOL NOT NULL DEFAULT true,
  status STRING NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_releases_project_created_at ON releases (project_id, created_at DESC);
CREATE INDEX idx_deployments_project_created_at ON deployments (project_id, created_at DESC);
CREATE INDEX idx_deployments_target_created_at ON deployments (runtime_target_id, created_at DESC);
CREATE INDEX idx_build_jobs_project_created_at ON build_jobs (project_id, created_at DESC);
CREATE INDEX idx_domains_project_service_id ON domains (project_service_id);
