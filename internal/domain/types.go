package domain

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

type RuntimeType string

const (
	RuntimeTypeDocker     RuntimeType = "docker"
	RuntimeTypeKubernetes RuntimeType = "kubernetes"
)

type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "pending"
	DeploymentStatusDeploying  DeploymentStatus = "deploying"
	DeploymentStatusHealthy    DeploymentStatus = "healthy"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back"
)

type BuildStatus string

const (
	BuildStatusPending   BuildStatus = "pending"
	BuildStatusRunning   BuildStatus = "running"
	BuildStatusSucceeded BuildStatus = "succeeded"
	BuildStatusFailed    BuildStatus = "failed"
)

type DomainStatus string

const (
	DomainStatusPending      DomainStatus = "pending"
	DomainStatusProvisioning DomainStatus = "provisioning"
	DomainStatusReady        DomainStatus = "ready"
	DomainStatusFailed       DomainStatus = "failed"
)
