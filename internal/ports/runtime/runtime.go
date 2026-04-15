package runtime

import "github.com/elouan/dockyard/internal/domain"

type DeploymentSpec struct {
	DeploymentID string
	ProjectID    string
	ProjectSlug  string
	ReleaseID    string
	Image        ImageRef
	Service      ServiceSpec
	Routing      *RoutingSpec
	Environment  []domain.EnvironmentVariable
	Strategy     string
}

type ImageRef struct {
	Repository string
	Tag        string
	Digest     string
}

type ServiceSpec struct {
	Name            string
	InternalPort    int
	HealthcheckPath string
	HealthcheckPort int
}

type RoutingSpec struct {
	Hostname    string
	Entrypoints []string
	TLS         bool
}

type DeploymentResult struct {
	DeploymentID string
	Status       domain.DeploymentStatus
	Message      string
	ContainerID  string
	StartedAt    string
	FinishedAt   string
}

type Driver interface {
	PrepareDeployment(spec DeploymentSpec) error
	ApplyRelease(spec DeploymentSpec) (DeploymentResult, error)
	CheckHealth(deploymentID string) (DeploymentResult, error)
	Rollback(deploymentID string, targetReleaseID string) (DeploymentResult, error)
	DeleteDeployment(deploymentID string) error
}
