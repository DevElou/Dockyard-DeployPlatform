package domain

import (
	"errors"
	"strings"
)

type Project struct {
	ID                   string        `json:"id"`
	Slug                 string        `json:"slug"`
	Name                 string        `json:"name"`
	Status               ProjectStatus `json:"status"`
	GitHubOwner          string        `json:"githubOwner"`
	GitHubRepo           string        `json:"githubRepo"`
	DefaultBranch        string        `json:"defaultBranch"`
	RootDirectory        string        `json:"rootDirectory"`
	DockerfilePath       string        `json:"dockerfilePath"`
	BuildContext         string        `json:"buildContext"`
	DefaultEnvironmentID string        `json:"defaultEnvironmentId"`
}

type RuntimeTarget struct {
	ID           string      `json:"id"`
	Slug         string      `json:"slug"`
	Name         string      `json:"name"`
	RuntimeType  RuntimeType `json:"runtimeType"`
	Endpoint     string      `json:"endpoint"`
	AgentKeyHash string      `json:"-"`
	ServerGroup  *string     `json:"serverGroup"`
	Region       *string     `json:"region"`
	Enabled      bool        `json:"enabled"`
}

type ProjectService struct {
	ID              string `json:"id"`
	ProjectID       string `json:"projectId"`
	Name            string `json:"name"`
	ContainerPort   int    `json:"containerPort"`
	HealthcheckPath string `json:"healthcheckPath"`
	HealthcheckPort int    `json:"healthcheckPort"`
	RoutingEnabled  bool   `json:"routingEnabled"`
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
