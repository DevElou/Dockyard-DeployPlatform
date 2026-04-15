package project

import (
	"context"
	"strings"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/repository"
)

type Service struct {
	projects repository.ProjectRepository
}

func NewService(projects repository.ProjectRepository) *Service {
	return &Service{projects: projects}
}

type CreateProjectInput struct {
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	GitHubOwner    string `json:"githubOwner"`
	GitHubRepo     string `json:"githubRepo"`
	DefaultBranch  string `json:"defaultBranch"`
	RootDirectory  string `json:"rootDirectory"`
	DockerfilePath string `json:"dockerfilePath"`
	BuildContext   string `json:"buildContext"`
}

func (s *Service) List(ctx context.Context) ([]domain.Project, error) {
	return s.projects.List(ctx)
}

func (s *Service) Create(ctx context.Context, input CreateProjectInput) (domain.Project, error) {
	project := domain.Project{
		Slug:           strings.TrimSpace(input.Slug),
		Name:           strings.TrimSpace(input.Name),
		GitHubOwner:    strings.TrimSpace(input.GitHubOwner),
		GitHubRepo:     strings.TrimSpace(input.GitHubRepo),
		DefaultBranch:  defaultString(input.DefaultBranch, "main"),
		RootDirectory:  defaultString(input.RootDirectory, "."),
		DockerfilePath: defaultString(input.DockerfilePath, "Dockerfile"),
		BuildContext:   defaultString(input.BuildContext, "."),
	}

	if err := project.Validate(); err != nil {
		return domain.Project{}, err
	}

	return s.projects.Create(ctx, project)
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
