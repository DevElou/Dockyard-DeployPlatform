package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/source"
)

type SourceProvider struct {
	token    string
	projects repository.ProjectRepository
	client   *http.Client
}

func NewSourceProvider(token string, projects repository.ProjectRepository) *SourceProvider {
	return &SourceProvider{
		token:    token,
		projects: projects,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *SourceProvider) ResolveRevision(ctx context.Context, projectID string, ref string) (source.Revision, error) {
	project, err := p.projects.GetByID(ctx, projectID)
	if err != nil {
		return source.Revision{}, fmt.Errorf("github: get project %s: %w", projectID, err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s",
		url.PathEscape(project.GitHubOwner),
		url.PathEscape(project.GitHubRepo),
		url.PathEscape(ref))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return source.Revision{}, fmt.Errorf("github: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := p.client.Do(req)
	if err != nil {
		return source.Revision{}, fmt.Errorf("github: request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return source.Revision{}, fmt.Errorf("github: ref %q not found in %s/%s",
			ref, project.GitHubOwner, project.GitHubRepo)

	case http.StatusUnauthorized, http.StatusForbidden:
		return source.Revision{}, fmt.Errorf("github: authentication failed (status %d)", resp.StatusCode)
	case http.StatusOK:
		// ok
	default:
		return source.Revision{}, fmt.Errorf("github: unexpected status %d for %s/%s@%s",
			resp.StatusCode, project.GitHubOwner, project.GitHubRepo, ref)
	}

	var commit struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return source.Revision{}, fmt.Errorf("github: decode response: %w", err)
	}

	return source.Revision{
		CommitSHA: commit.SHA,
		GitRef:    ref,
		ArchiveURL: fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s",
			project.GitHubOwner, project.GitHubRepo, commit.SHA),
	}, nil
}
