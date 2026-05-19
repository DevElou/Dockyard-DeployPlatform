package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/elouan/dockyard/internal/domain"
)

var errNotImplemented = errors.New("not implemented in memory adapter")

type ProjectRepository struct {
	mu       sync.RWMutex
	sequence int
	projects []domain.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make([]domain.Project, 0),
	}
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Project, len(r.projects))
	copy(result, r.projects)
	return result, nil
}

func (r *ProjectRepository) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sequence++
	project.ID = fmt.Sprintf("proj_%04d", r.sequence)
	r.projects = append(r.projects, project)

	return project, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (domain.Project, error) {
	return domain.Project{}, errNotImplemented
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	return domain.Project{}, errNotImplemented
}

func (r *ProjectRepository) Archive(ctx context.Context, id string) error {
	return errNotImplemented
}

func (r *ProjectRepository) ListRuntimeTargets(ctx context.Context, projectID string) ([]domain.RuntimeTarget, error) {
	return nil, errNotImplemented
}

func (r *ProjectRepository) AddRuntimeTarget(ctx context.Context, projectID, runtimeTargetID string) error {
	return errNotImplemented
}
