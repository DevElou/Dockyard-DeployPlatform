package domain

type Deployment struct {
	ID                     string
	ProjectID              string
	ReleaseID              string
	RuntimeTargetID        string
	ProjectServiceID       string
	EnvironmentSetID       string
	Status                 DeploymentStatus
	Strategy               string
	RollbackOfDeploymentID string
	CreatedAt              string
}

type Domain struct {
	ID               string
	ProjectID        string
	ProjectServiceID string
	Hostname         string
	BaseDomain       string
	Provider         string
	RoutingType      string
	Status           DomainStatus
}

type EnvironmentVariable struct {
	Key      string
	Value    string
	IsSecret bool
}
