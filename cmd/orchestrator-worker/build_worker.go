package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/registry"
	"github.com/elouan/dockyard/internal/ports/repository"
	"github.com/elouan/dockyard/internal/ports/source"
)

type BuildWorker struct {
	releases repository.ReleaseRepository
	projects repository.ProjectRepository
	source   source.Provider
	builder  registry.Builder
	inFlight sync.Map // releaseID → struct{}
}

func NewBuildWorker(
	releases repository.ReleaseRepository,
	projects repository.ProjectRepository,
	src source.Provider,
	builder registry.Builder,
) *BuildWorker {
	return &BuildWorker{
		releases: releases,
		projects: projects,
		source:   src,
		builder:  builder,
	}
}

func (w *BuildWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("build-worker: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("build-worker: stopping")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *BuildWorker) tick(ctx context.Context) {
	pending, err := w.releases.ListByBuildStatus(ctx, domain.BuildStatusPending)
	if err != nil {
		log.Printf("build-worker: list pending releases: %v", err)
		return
	}

	for _, r := range pending {
		if _, loaded := w.inFlight.LoadOrStore(r.ID, struct{}{}); loaded {
			continue
		}
		go func(rel domain.Release) {
			defer w.inFlight.Delete(rel.ID)
			w.processRelease(ctx, rel)
		}(r)
	}
}

func (w *BuildWorker) processRelease(ctx context.Context, r domain.Release) {
	log.Printf("build-worker: processing release %s", r.ID)

	if err := w.releases.UpdateBuildStatus(ctx, r.ID, domain.BuildStatusRunning); err != nil {
		log.Printf("build-worker: claim release %s: %v", r.ID, err)
		return
	}

	result, err := w.build(ctx, r)
	if err != nil {
		log.Printf("build-worker: build release %s: %v", r.ID, err)
		w.failBuild(r.ID)
		return
	}

	if err := w.releases.UpdateBuildResult(ctx, r.ID, result.ImageRepository, result.ImageTag, result.ImageDigest, domain.BuildStatusSucceeded); err != nil {
		log.Printf("build-worker: persist build result for release %s: %v", r.ID, err)
	}
	log.Printf("build-worker: release %s built successfully: %s@%s", r.ID, result.ImageTag, result.ImageDigest)
}

func (w *BuildWorker) build(ctx context.Context, r domain.Release) (registry.BuildResult, error) {
	project, err := w.projects.GetByID(ctx, r.ProjectID)
	if err != nil {
		return registry.BuildResult{}, fmt.Errorf("get project: %w", err)
	}

	workDir, err := os.MkdirTemp("", fmt.Sprintf("dockyard-build-%s-*", r.ID[:8]))
	if err != nil {
		return registry.BuildResult{}, fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	if err := w.source.DownloadArchive(ctx, r.ProjectID, r.GitSHA, workDir); err != nil {
		return registry.BuildResult{}, fmt.Errorf("download archive: %w", err)
	}

	dockerfilePath := filepath.Join(workDir, project.DockerfilePath)

	return w.builder.BuildAndPush(ctx, registry.BuildRequest{
		ProjectID:      r.ProjectID,
		ReleaseVersion: r.Version,
		CommitSHA:      r.GitSHA,
		BuildContext:   workDir,
		DockerfilePath: dockerfilePath,
	})
}

func (w *BuildWorker) failBuild(releaseID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.releases.UpdateBuildStatus(ctx, releaseID, domain.BuildStatusFailed); err != nil {
		log.Printf("build-worker: fail release %s: %v", releaseID, err)
	}
}
