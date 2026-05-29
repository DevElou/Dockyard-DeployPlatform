package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elouan/dockyard/internal/application/operationlog"
	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/registry"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/source"
)

// Release build phases — kept narrow so the UI timeline reads naturally.
const (
	phaseBuildQueued            = "queued"
	phaseBuildDownloadingSource = "downloading_archive"
	phaseBuildBuildingImage     = "building_image"
	phaseBuildSucceeded         = "succeeded"
	phaseBuildFailed            = "failed"
)

type BuildWorker struct {
	releases repository.ReleaseRepository
	projects repository.ProjectRepository
	source   source.Provider
	builder  registry.Builder
	events   *operationlog.Service
	inFlight sync.Map // releaseID → struct{}
}

func NewBuildWorker(
	releases repository.ReleaseRepository,
	projects repository.ProjectRepository,
	src source.Provider,
	builder registry.Builder,
	events *operationlog.Service,
) *BuildWorker {
	return &BuildWorker{
		releases: releases,
		projects: projects,
		source:   src,
		builder:  builder,
		events:   events,
	}
}

func (w *BuildWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var wg sync.WaitGroup
	log.Println("build-worker: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("build-worker: stopping, draining in-flight builds")
			wg.Wait()
			return
		case <-ticker.C:
			w.tick(ctx, &wg)
		}
	}
}

func (w *BuildWorker) tick(ctx context.Context, wg *sync.WaitGroup) {
	pending, err := w.releases.ListByBuildStatus(ctx, domain.BuildStatusPending)
	if err != nil {
		log.Printf("build-worker: list pending releases: %v", err)
		return
	}

	for _, r := range pending {
		if _, loaded := w.inFlight.LoadOrStore(r.ID, struct{}{}); loaded {
			continue
		}
		wg.Add(1)
		go func(rel domain.Release) {
			defer wg.Done()
			defer w.inFlight.Delete(rel.ID)
			w.processRelease(ctx, rel)
		}(r)
	}
}

func (w *BuildWorker) processRelease(ctx context.Context, r domain.Release) {
	log.Printf("build-worker: processing release %s", r.ID)
	w.events.Info(ctx, domain.OperationResourceRelease, r.ID, phaseBuildQueued,
		"build worker picked up release", nil)

	if err := w.releases.UpdateBuildStatus(ctx, r.ID, domain.BuildStatusRunning); err != nil {
		log.Printf("build-worker: claim release %s: %v", r.ID, err)
		w.events.Error(ctx, domain.OperationResourceRelease, r.ID, phaseBuildFailed,
			"failed to mark build as running", map[string]string{"error": err.Error()})
		return
	}

	result, err := w.build(ctx, r)
	if err != nil {
		log.Printf("build-worker: build release %s: %v", r.ID, err)
		w.events.Error(ctx, domain.OperationResourceRelease, r.ID, phaseBuildFailed,
			"build failed", map[string]string{"error": truncate(err.Error(), 4000)})
		w.failBuild(r.ID)
		return
	}

	if err := w.releases.UpdateBuildResult(ctx, r.ID, result.ImageRepository, result.ImageTag, result.ImageDigest, domain.BuildStatusSucceeded); err != nil {
		log.Printf("build-worker: persist build result for release %s: %v", r.ID, err)
		w.events.Error(ctx, domain.OperationResourceRelease, r.ID, phaseBuildFailed,
			"failed to persist build result", map[string]string{"error": err.Error()})
		return
	}
	log.Printf("build-worker: release %s built successfully: %s@%s", r.ID, result.ImageTag, result.ImageDigest)
	w.events.Success(ctx, domain.OperationResourceRelease, r.ID, phaseBuildSucceeded,
		"image built and pushed",
		map[string]string{
			"imageRepository": result.ImageRepository,
			"imageTag":        result.ImageTag,
			"imageDigest":     result.ImageDigest,
		})
}

func (w *BuildWorker) build(ctx context.Context, r domain.Release) (registry.BuildResult, error) {
	project, err := w.projects.GetByID(ctx, r.ProjectID)
	if err != nil {
		return registry.BuildResult{}, fmt.Errorf("get project: %w", err)
	}

	prefix := r.ID
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	workDir, err := os.MkdirTemp("", fmt.Sprintf("dockyard-build-%s-*", prefix))
	if err != nil {
		return registry.BuildResult{}, fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	w.events.Info(ctx, domain.OperationResourceRelease, r.ID, phaseBuildDownloadingSource,
		"downloading source archive from GitHub",
		map[string]string{"gitSha": r.GitSHA, "gitRef": r.GitRef})

	if err := w.source.DownloadArchive(ctx, r.ProjectID, r.GitSHA, workDir); err != nil {
		return registry.BuildResult{}, fmt.Errorf("download archive: %w", err)
	}

	buildContext, dockerfilePath, err := resolveBuildPaths(workDir, project)
	if err != nil {
		return registry.BuildResult{}, fmt.Errorf("resolve build paths: %w", err)
	}

	w.events.Info(ctx, domain.OperationResourceRelease, r.ID, phaseBuildBuildingImage,
		"building and pushing Docker image",
		map[string]string{
			"rootDirectory":  project.RootDirectory,
			"buildContext":   project.BuildContext,
			"dockerfilePath": project.DockerfilePath,
		})

	return w.builder.BuildAndPush(ctx, registry.BuildRequest{
		ProjectID:      r.ProjectID,
		ReleaseVersion: r.Version,
		CommitSHA:      r.GitSHA,
		BuildContext:   buildContext,
		DockerfilePath: dockerfilePath,
	})
}

func resolveBuildPaths(workDir string, project domain.Project) (string, string, error) {
	rootDir, err := safeJoin(workDir, defaultBuildPath(project.RootDirectory))
	if err != nil {
		return "", "", fmt.Errorf("root directory: %w", err)
	}

	buildContext, err := safeJoin(rootDir, defaultBuildPath(project.BuildContext))
	if err != nil {
		return "", "", fmt.Errorf("build context: %w", err)
	}

	if st, err := os.Stat(buildContext); err != nil {
		return "", "", fmt.Errorf("build context %q: %w", project.BuildContext, err)
	} else if !st.IsDir() {
		return "", "", fmt.Errorf("build context %q is not a directory", project.BuildContext)
	}

	dockerfilePath, tried, err := resolveDockerfilePath(rootDir, buildContext, defaultBuildPath(project.DockerfilePath))
	if err != nil {
		return "", "", fmt.Errorf("dockerfile %q not found (tried: %s)", project.DockerfilePath, strings.Join(tried, ", "))
	}

	return buildContext, dockerfilePath, nil
}

func resolveDockerfilePath(rootDir, buildContext, dockerfileRel string) (string, []string, error) {
	var tried []string
	for _, base := range uniquePaths(buildContext, rootDir) {
		candidate, err := safeJoin(base, dockerfileRel)
		if err != nil {
			return "", tried, err
		}
		tried = append(tried, candidate)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, tried, nil
		}
	}
	return "", tried, os.ErrNotExist
}

func uniquePaths(paths ...string) []string {
	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, p := range paths {
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		unique = append(unique, p)
	}
	return unique
}

func defaultBuildPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "."
	}
	return value
}

// safeJoin joins base and rel, returning an error if the result escapes base.
func safeJoin(base, rel string) (string, error) {
	rel = filepath.Clean(filepath.FromSlash(strings.TrimSpace(rel)))
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q must be relative", rel)
	}

	joined := filepath.Clean(filepath.Join(base, rel))
	cleanBase := filepath.Clean(base)
	if joined != cleanBase && !strings.HasPrefix(joined, cleanBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes work directory", rel)
	}
	return joined, nil
}

func (w *BuildWorker) failBuild(releaseID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.releases.UpdateBuildStatus(ctx, releaseID, domain.BuildStatusFailed); err != nil {
		log.Printf("build-worker: fail release %s: %v", releaseID, err)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
