package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/elouan/dockyard/internal/domain"
)

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
