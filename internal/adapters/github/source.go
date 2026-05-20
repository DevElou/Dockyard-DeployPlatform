package github

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
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

// maxArchiveFileBytes is the per-entry size cap during tar extraction (500 MB).
const maxArchiveFileBytes int64 = 500 << 20

type SourceProvider struct {
	token         string
	projects      repository.ProjectRepository
	client        *http.Client // API calls (short timeout)
	archiveClient *http.Client // tarball downloads (long timeout)
}

func NewSourceProvider(token string, projects repository.ProjectRepository) *SourceProvider {
	return &SourceProvider{
		token:         token,
		projects:      projects,
		client:        &http.Client{Timeout: 15 * time.Second},
		archiveClient: &http.Client{Timeout: 5 * time.Minute},
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return fmt.Errorf("github: build archive request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := p.archiveClient.Do(req)
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
// Symlinks and hard-links are rejected to prevent post-extraction traversal.
// Each regular file is capped at maxArchiveFileBytes to prevent decompression bombs.
func extractTarGz(r io.Reader, dir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("github: open gzip: %w", err)
	}
	defer gz.Close()

	safeDir := filepath.Clean(dir) + string(os.PathSeparator)

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("github: read tar: %w", err)
		}

		stripped := stripFirstComponent(hdr.Name)
		if stripped == "" {
			continue
		}

		target := filepath.Join(dir, filepath.FromSlash(stripped))

		if !strings.HasPrefix(target, safeDir) {
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
			if err := writeFile(target, tr, hdr.FileInfo().Mode()); err != nil {
				return err
			}

		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("github: symlinks not permitted in archive: %s", hdr.Name)

		default:
			// Skip device files, FIFOs, etc.
		}
	}
	return nil
}

func writeFile(target string, src io.Reader, mode os.FileMode) error {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("github: create file %s: %w", target, err)
	}

	// CopyN with maxArchiveFileBytes+1: if it returns nil, the file exceeded the limit.
	_, copyErr := io.CopyN(f, src, maxArchiveFileBytes+1)
	f.Close()

	if copyErr == nil {
		// Exactly maxArchiveFileBytes+1 bytes copied — file is too large.
		os.Remove(target)
		return fmt.Errorf("github: file %s exceeds %d byte limit", target, maxArchiveFileBytes)
	}
	if !errors.Is(copyErr, io.EOF) {
		return fmt.Errorf("github: write file %s: %w", target, copyErr)
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
