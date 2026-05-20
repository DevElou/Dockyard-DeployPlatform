package domain

import (
	"errors"
	"strings"
)

type Project struct {
	ID                   string
	Slug                 string
	Name                 string
	Status               ProjectStatus
	GitHubOwner          string
	GitHubRepo           string
	DefaultBranch        string
	RootDirectory        string
	DockerfilePath       string
	BuildContext         string
	DefaultEnvironmentID string
}

type RuntimeTarget struct {
	ID           string
	Slug         string
	Name         string
	RuntimeType  RuntimeType
	Endpoint     string
	AgentKeyHash string
	ServerGroup  *string
	Region       *string
	Enabled      bool
}

type ProjectService struct {
	ID              string
	ProjectID       string
	Name            string
	ContainerPort   int
	HealthcheckPath string
	HealthcheckPort int
	RoutingEnabled  bool
}

func (p Project) Validate() error {
	if strings.TrimSpace(p.Slug) == "" {
		return errors.New("project slug is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("project name is required")
	}
	if strings.TrimSpace(p.GitHubOwner) == "" {
		return errors.New("github owner is required")
	}
	if strings.TrimSpace(p.GitHubRepo) == "" {
		return errors.New("github repo is required")
	}
	return nil
}

func (rt RuntimeTarget) Validate() error {
	if strings.TrimSpace(rt.Slug) == "" {
		return errors.New("runtime target slug is required")
	}
	if strings.TrimSpace(rt.Name) == "" {
		return errors.New("runtime target name is required")
	}
	if strings.TrimSpace(rt.Endpoint) == "" {
		return errors.New("runtime target endpoint is required")
	}
	if strings.TrimSpace(rt.AgentKeyHash) == "" {
		return errors.New("runtime target agent key hash is required")
	}
	return nil
}
