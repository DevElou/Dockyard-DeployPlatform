package domain

import (
	"errors"
	"strings"
	"time"
)

type Deployment struct {
	ID                     string           `json:"id"`
	ProjectID              string           `json:"projectId"`
	ReleaseID              string           `json:"releaseId"`
	RuntimeTargetID        string           `json:"runtimeTargetId"`
	ProjectServiceID       *string          `json:"projectServiceId,omitempty"`
	EnvironmentSetID       *string          `json:"environmentSetId,omitempty"`
	Status                 DeploymentStatus `json:"status"`
	Strategy               string           `json:"strategy"`
	TriggeredByUserID      *string          `json:"triggeredByUserId,omitempty"`
	RollbackOfDeploymentID *string          `json:"rollbackOfDeploymentId,omitempty"`
	StartedAt              *time.Time       `json:"startedAt,omitempty"`
	FinishedAt             *time.Time       `json:"finishedAt,omitempty"`
	CreatedAt              time.Time        `json:"createdAt"`
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
	ID               string       `json:"id"`
	ProjectID        string       `json:"projectId"`
	ProjectServiceID *string      `json:"projectServiceId,omitempty"`
	Hostname         string       `json:"hostname"`
	BaseDomain       string       `json:"baseDomain"`
	Provider         string       `json:"provider"`
	RoutingType      string       `json:"routingType"`
	TLSEnabled       bool         `json:"tlsEnabled"`
	Status           DomainStatus `json:"status"`
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
	ID               string `json:"id"`
	EnvironmentSetID string `json:"environmentSetId"`
	Key              string `json:"key"`
	Value            string `json:"value"`
	IsSecret         bool   `json:"isSecret"`
}

type EnvironmentSet struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
}
