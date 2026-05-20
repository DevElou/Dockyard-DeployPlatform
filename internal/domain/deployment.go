package domain

import (
	"errors"
	"strings"
	"time"
)

type Deployment struct {
	ID                     string
	ProjectID              string
	ReleaseID              string
	RuntimeTargetID        string
	ProjectServiceID       *string
	EnvironmentSetID       *string
	Status                 DeploymentStatus
	Strategy               string
	TriggeredByUserID      *string
	RollbackOfDeploymentID *string
	StartedAt              *time.Time
	FinishedAt             *time.Time
	CreatedAt              time.Time
}

func (d Deployment) Validate() error {
	if strings.TrimSpace(d.ProjectID) == "" {
		return errors.New("deployment project ID is required")
	}
	if strings.TrimSpace(d.ReleaseID) == "" {
		return errors.New("deployment release ID is required")
	}
	if strings.TrimSpace(d.RuntimeTargetID) == "" {
		return errors.New("deployment runtime target ID is required")
	}
	return nil
}

type Domain struct {
	ID               string
	ProjectID        string
	ProjectServiceID *string
	Hostname         string
	BaseDomain       string
	Provider         string
	RoutingType      string
	TLSEnabled       bool
	Status           DomainStatus
}

func (d Domain) Validate() error {
	if strings.TrimSpace(d.ProjectID) == "" {
		return errors.New("domain project ID is required")
	}
	if strings.TrimSpace(d.Hostname) == "" {
		return errors.New("domain hostname is required")
	}
	if strings.TrimSpace(d.BaseDomain) == "" {
		return errors.New("domain base domain is required")
	}
	return nil
}

type EnvironmentVariable struct {
	ID               string
	EnvironmentSetID string
	Key              string
	Value            string
	IsSecret         bool
}

type EnvironmentSet struct {
	ID        string
	ProjectID string
	Name      string
}
