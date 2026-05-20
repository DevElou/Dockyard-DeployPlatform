package github

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

// DownloadArchive fetches the GitHub tarball for commitSHA and extracts it into
// targetDir, stripping the top-level directory GitHub prepends (owner-repo-sha/).
func (p *SourceProvider) DownloadArchive(ctx context.Context, projectID string, commitSHA string, targetDir string) error {
	project, err := p.projects.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("github: get project %s: %w", projectID, err)
	}

	archiveURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s",
		url.PathEscape(project.GitHubOwner),
		url.PathEscape(project.GitHubRepo),
		url.PathEscape(commitSHA))

	// GitHub tarballs can redirect; use a client with a generous timeout.
	dlClient := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return fmt.Errorf("github: build archive request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := dlClient.Do(req)
	if err != nil {
		return fmt.Errorf("github: download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github: download archive status %d", resp.StatusCode)
	}

	return extractTarGz(resp.Body, targetDir)
}

// extractTarGz extracts a gzip-compressed tar stream into dir, stripping the
// first path component (the "owner-repo-sha/" prefix that GitHub adds).
func extractTarGz(r io.Reader, dir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("github: open gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("github: read tar: %w", err)
		}

		// Strip the first component (e.g. "owner-repo-sha/").
		stripped := stripFirstComponent(hdr.Name)
		if stripped == "" {
			continue
		}

		target := filepath.Join(dir, filepath.FromSlash(stripped))

		// Guard against path traversal.
		if !strings.HasPrefix(target, filepath.Clean(dir)+string(os.PathSeparator)) {
			return fmt.Errorf("github: unsafe path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("github: mkdir %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("github: mkdir parent %s: %w", target, err)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("github: create file %s: %w", target, err)
			}
			if _, err := io.Copy(f, tr); err != nil { //nolint:gosec
				f.Close()
				return fmt.Errorf("github: write file %s: %w", target, err)
			}
			f.Close()
		}
	}
	return nil
}

func stripFirstComponent(p string) string {
	p = filepath.ToSlash(p)
	idx := strings.Index(p, "/")
	if idx < 0 {
		return ""
	}
	return p[idx+1:]
}
